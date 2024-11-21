package sql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/dicedb/dice/internal/sql/parser"
)

// Constants for custom syntax replacements
const (
	CustomKey   = "$key"
	CustomValue = "$value"
	TempPrefix  = "_"
	TempKey     = TempPrefix + "key"
	TempValue   = TempPrefix + "value"
)

type ComparisonOp string

const (
	OpEq      ComparisonOp = "="
	OpNeq     ComparisonOp = "!="
	OpGt      ComparisonOp = ">"
	OpGte     ComparisonOp = ">="
	OpLt      ComparisonOp = "<"
	OpLte     ComparisonOp = "<="
	OpLike    ComparisonOp = "like"
	OpNotLike ComparisonOp = "notlike"
)

type FieldType string

const (
	FieldJSON   FieldType = "json"
	FieldKey    FieldType = "key"
	FieldVal    FieldType = "value"
	FieldInt    FieldType = "int64"
	FieldString FieldType = "string"
	FieldFloat  FieldType = "float"
	FieldOth    FieldType = "other"
)

type Value struct {
	Type  FieldType
	Value string
}

func NewJSONType(val string) Value {
	return Value{
		Value: val,
		Type:  FieldJSON,
	}
}

func NewIntType(val string) Value {
	return Value{
		Value: val,
		Type:  FieldInt,
	}
}

func NewStringType(val string) Value {
	return Value{
		Value: val,
		Type:  FieldString,
	}
}

func NewFloatType(val string) Value {
	return Value{
		Value: val,
		Type:  FieldFloat,
	}
}

func NewKeyType() Value {
	return Value{
		Value: TempKey,
		Type:  FieldKey,
	}
}

func NewValueType() Value {
	return Value{
		Value: TempValue,
		Type:  FieldVal,
	}
}

// ConditionNode represents any node in the condition tree.
type ConditionNode interface {
	isConditionNode()
}

type AndNode struct {
	Left  ConditionNode
	Right ConditionNode
}

func (a *AndNode) isConditionNode() {}

type OrNode struct {
	Left  ConditionNode
	Right ConditionNode
}

func (o *OrNode) isConditionNode() {}

type ComparisonNode struct {
	Left     Value        // _key or _value
	Operator ComparisonOp // Comparison operator
	Right    Value        // Resolved value to compare with
}

func (c *ComparisonNode) isConditionNode() {}

type QuerySelection struct {
	KeySelection   bool
	ValueSelection bool
}

type QueryOrder struct {
	OrderBy Value
	Order   string
}

type DSQLQuery struct {
	Selection   QuerySelection
	Where       ConditionNode
	OrderBy     QueryOrder
	Limit       int
	Fingerprint string
}

type dsqlListner struct {
	*parser.BasedsqlListener
	dsqlQuery *DSQLQuery
}

type dsqlErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []error
}

func newDsqlErrorListner() *dsqlErrorListener {
	return &dsqlErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
	}
}

func (l *dsqlErrorListener) SyntaxError(_ antlr.Recognizer, _ interface{}, line, col int, msg string, _ antlr.RecognitionException) {
	l.Errors = append(l.Errors, fmt.Errorf("syntax error at line %d:%d - %s", line, col, msg))
}

func (l *dsqlListner) ExitSelectStmt(ctx *parser.SelectStmtContext) {
	l.dsqlQuery.Selection = parseSelectExpressions(ctx.SelectFields())
}

func (l *dsqlListner) ExitOrderByClause(ctx *parser.OrderByClauseContext) {
	l.dsqlQuery.OrderBy, _ = parseOrderBy(ctx)
}

func (l *dsqlListner) ExitLimitClause(ctx *parser.LimitClauseContext) {
	l.dsqlQuery.Limit = parseLimit(ctx)
}

func (l *dsqlListner) ExitWhereClause(ctx *parser.WhereClauseContext) {
	l.dsqlQuery.Where, _ = ParseWhereClause(ctx)
}

// replacePlaceholders replaces temporary placeholders with custom ones
func replacePlaceholders(s string) string {
	replacer := strings.NewReplacer(
		TempKey, CustomKey,
		TempValue, CustomValue,
	)
	return replacer.Replace(s)
}

func (q DSQLQuery) String() string {
	var parts []string

	// Selection
	var selectionParts []string
	if q.Selection.KeySelection {
		selectionParts = append(selectionParts, CustomKey)
	}
	if q.Selection.ValueSelection {
		selectionParts = append(selectionParts, CustomValue)
	}
	if len(selectionParts) > 0 {
		parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(selectionParts, ", ")))
	} else {
		parts = append(parts, "SELECT *")
	}

	// Where
	if q.Where != nil {
		whereClause := whereClauseToSQL(q.Where)
		whereClause = replacePlaceholders(whereClause)
		parts = append(parts, fmt.Sprintf("WHERE %s", whereClause))
	}

	// OrderBy
	if q.OrderBy.OrderBy.Value != "" {
		orderByClause := replacePlaceholders(q.OrderBy.OrderBy.Value)
		parts = append(parts, fmt.Sprintf("ORDER BY %s %s", orderByClause, q.OrderBy.Order))
	}

	// Limit
	if q.Limit > 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d", q.Limit))
	}

	return strings.Join(parts, " ")
}

// Utility functions for custom syntax handling
func replaceCustomSyntax(sql string) string {
	replacer := strings.NewReplacer(CustomKey, TempKey, CustomValue, TempValue)

	// Add implicit `FROM dual` if no table name is provided
	return replacer.Replace(sql)
}

// ParseQuery takes a SQL query string and returns a DSQLQuery struct
func ParseQuery(sql string) (*DSQLQuery, error) {
	sql = replaceCustomSyntax(sql)

	is := antlr.NewInputStream(sql)

	errorListener := newDsqlErrorListner()

	lexer := parser.NewdsqlLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errorListener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewdsqlParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(errorListener)

	listner := &dsqlListner{
		dsqlQuery: &DSQLQuery{},
	}
	antlr.ParseTreeWalkerDefault.Walk(listner, p.Query())

	if len(errorListener.Errors) > 0 {
		return nil, fmt.Errorf("error parsing SQL statement: %v", errorListener.Errors)
	}

	listner.dsqlQuery.Fingerprint = generateFingerprint(listner.dsqlQuery.Where)
	return listner.dsqlQuery, nil
}

func parseSelectExpressions(ctx parser.ISelectFieldsContext) QuerySelection {
	selection := QuerySelection{}

	if ctx == nil {
		return selection
	}

	if ctx.STAR() != nil {
		selection.KeySelection = true
		selection.ValueSelection = true
		return selection
	}

	for _, field := range ctx.AllField() {
		switch field.GetText() {
		case TempKey:
			selection.KeySelection = true
		case TempValue:
			selection.ValueSelection = true
		}
	}

	return selection
}

func parseOrderBy(ctx parser.IOrderByClauseContext) (QueryOrder, error) {
	if ctx == nil {
		return QueryOrder{}, nil
	}

	orderBy := Value{}
	orderBy.Value = trimQuotesOrBackticks(ctx.OrderByField().FieldWithString().GetText())
	orderBy.Type = getType(orderBy.Value)

	if orderBy.Type == FieldInt || orderBy.Type == FieldFloat {
		return QueryOrder{}, fmt.Errorf("invalid order by field type %s", orderBy.Type)
	}

	if ctx.OrderByField().DESC() != nil {
		return QueryOrder{OrderBy: orderBy, Order: Desc}, nil
	}
	return QueryOrder{OrderBy: orderBy, Order: Asc}, nil
}

func parseLimit(ctx parser.ILimitClauseContext) int {
	if ctx != nil {
		if ctx.NUMBER() != nil {
			limitStr := ctx.NUMBER().GetText()
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				panic("error in grammar")
			}
			return limit
		}
	}
	return 0
}

// ParseWhereClause parses and returns the root of the condition tree.
func ParseWhereClause(ctx parser.IWhereClauseContext) (ConditionNode, error) {
	root, err := buildConditionNode(ctx.Condition())
	if err != nil {
		return nil, err
	}
	return compact(root), nil
}

// Recursive function to build the condition AST with parentheses support.
func buildConditionNode(ctx parser.IConditionContext) (ConditionNode, error) {
	if ctx.LPAREN() != nil && ctx.RPAREN() != nil {
		innerCtx := ctx.GetChild(1).(*parser.ConditionContext)
		return buildConditionNode(innerCtx)
	}

	if ctx.AND() != nil || ctx.OR() != nil {
		// Parse both sides of the logical expression
		leftExpr := ctx.GetChild(0).(*parser.ConditionContext)
		rightExpr := ctx.GetChild(2).(*parser.ConditionContext)

		leftNode, err := buildConditionNode(leftExpr)
		if err != nil {
			return nil, err
		}
		rightNode, err := buildConditionNode(rightExpr)
		if err != nil {
			return nil, err
		}

		if ctx.AND() != nil {
			return &AndNode{Left: leftNode, Right: rightNode}, nil
		}
		return &OrNode{Left: leftNode, Right: rightNode}, nil
	}

	// Handle leaf nodes with comparison expressions
	if ctx.Expression() != nil {
		expr := ctx.Expression().(*parser.ExpressionContext)
		field := trimQuotesOrBackticks(expr.FieldWithString().GetText())
		operator := ComparisonOp(strings.ToLower(expr.ComparisonOp().GetText()))
		value := trimQuotesOrBackticks(expr.Value().GetText())

		return &ComparisonNode{
			Left:     Value{Value: field, Type: getType(field)},
			Operator: operator,
			Right:    Value{Value: value, Type: getType(value)},
		}, nil
	}
	return nil, fmt.Errorf("invalid condition context")
}

func compact(node ConditionNode) ConditionNode {
	switch n := node.(type) {
	case *ComparisonNode:
		return n // Leaf node, return as-is
	case *AndNode:
		left := compact(n.Left)
		right := compact(n.Right)
		if isEqualNode(left, right) {
			return left
		}
		return &AndNode{Left: left, Right: right}
	case *OrNode:
		left := compact(n.Left)
		right := compact(n.Right)
		if isEqualNode(left, right) {
			return left
		}
		return &OrNode{Left: left, Right: right}
	}
	return node
}

func isEqualNode(a, b ConditionNode) bool {
	switch a := a.(type) {
	case *ComparisonNode:
		b, ok := b.(*ComparisonNode)
		return ok && a == b
	case *AndNode:
		b, ok := b.(*AndNode)
		return ok && isEqualNode(a.Left, b.Left) && isEqualNode(a.Right, b.Right)
	case *OrNode:
		b, ok := b.(*OrNode)
		return ok && isEqualNode(a.Left, b.Left) && isEqualNode(a.Right, b.Right)
	default:
		return false
	}
}

func getType(value string) FieldType {
	switch value {
	case TempKey:
		return FieldKey
	case TempValue:
		return FieldVal
	default:
		if strings.HasPrefix(value, TempPrefix) {
			return FieldJSON
		}

		if _, err := strconv.Atoi(value); err == nil {
			return FieldInt
		}

		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return FieldFloat
		}

		return FieldString
	}
}

// Helper function to trim both single and double quotes/backticks
func trimQuotesOrBackticks(input string) string {
	if len(input) > 1 && ((input[0] == '\'' && input[len(input)-1] == '\'') ||
		(input[0] == '`' && input[len(input)-1] == '`')) {
		return input[1 : len(input)-1]
	}
	return input
}

// Helper function to convert ConditionNode to SQL
func conditionNodeToSQL(node ConditionNode) string {
	if node == nil {
		return ""
	}

	// Check if the node is a logical operation
	switch node := node.(type) {
	case *ComparisonNode:
		return comparisonExprToSQL(node)
	case *OrNode:
		leftSQL := conditionNodeToSQL(node.Left)
		rightSQL := conditionNodeToSQL(node.Right)
		return fmt.Sprintf("%s %s %s", leftSQL, "OR", rightSQL)
	case *AndNode:
		leftSQL := conditionNodeToSQL(node.Left)
		rightSQL := conditionNodeToSQL(node.Right)
		return fmt.Sprintf("%s %s %s", leftSQL, "AND", rightSQL)
	default:
		return ""
	}
}

// Helper function to convert ComparisonExpr to SQL
func comparisonExprToSQL(expr *ComparisonNode) string {
	sqlBuilder := &strings.Builder{}
	sqlBuilder.WriteString(expr.Left.Value)
	sqlBuilder.WriteString(" ")
	sqlBuilder.WriteString(string(expr.Operator))
	sqlBuilder.WriteString(" ")
	sqlBuilder.WriteString(expr.Right.Value)

	return sqlBuilder.String()
}

func whereClauseToSQL(node ConditionNode) string {
	return conditionNodeToSQL(node)
}
