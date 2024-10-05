package sql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xwb1989/sqlparser"
)

// Constants for custom syntax replacements
const (
	CustomKey   = "$key"
	CustomValue = "$value"
	TempPrefix  = "_"
	TempKey     = TempPrefix + "key"
	TempValue   = TempPrefix + "value"
)

// UnsupportedDSQLStatementError is returned when a DSQL statement is not supported
type UnsupportedDSQLStatementError struct {
	Stmt sqlparser.Statement
}

func (e *UnsupportedDSQLStatementError) Error() string {
	return fmt.Sprintf("unsupported DSQL statement: %T", e.Stmt)
}

func newUnsupportedSQLStatementError(stmt sqlparser.Statement) *UnsupportedDSQLStatementError {
	return &UnsupportedDSQLStatementError{Stmt: stmt}
}

// QuerySelection represents the SELECT expressions in the query
type QuerySelection struct {
	KeySelection   bool
	ValueSelection bool
}

type QueryOrder struct {
	OrderBy string
	Order   string
}

type DSQLQuery struct {
	Selection   QuerySelection
	Where       sqlparser.Expr
	OrderBy     QueryOrder
	Limit       int
	Fingerprint string
}

// replacePlaceholders replaces temporary placeholders with custom ones
func replacePlaceholders(s string) string {
	replacer := strings.NewReplacer(
		TempKey, CustomKey,
		TempValue, CustomValue,
	)
	return replacer.Replace(s)
}

// parseSelectExpressions parses the SELECT expressions in the query

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
		whereClause := sqlparser.String(q.Where)
		whereClause = replacePlaceholders(whereClause)
		parts = append(parts, fmt.Sprintf("WHERE %s", whereClause))
	}

	// OrderBy
	if q.OrderBy.OrderBy != "" {
		orderByClause := replacePlaceholders(q.OrderBy.OrderBy)
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
func ParseQuery(sql string) (DSQLQuery, error) {
	// Replace custom syntax before parsing
	sql = replaceCustomSyntax(sql)

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return DSQLQuery{}, fmt.Errorf("error parsing SQL statement: %v", err)
	}

	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		return DSQLQuery{}, newUnsupportedSQLStatementError(stmt)
	}

	// Ensure no unsupported clauses are present
	if err := checkUnsupportedClauses(selectStmt); err != nil {
		return DSQLQuery{}, err
	}

	querySelection, err := parseSelectExpressions(selectStmt)
	if err != nil {
		return DSQLQuery{}, err
	}

	if err := parseTableName(selectStmt); err != nil {
		return DSQLQuery{}, err
	}

	where := parseWhere(selectStmt)

	orderBy, err := parseOrderBy(selectStmt)
	if err != nil {
		return DSQLQuery{}, err
	}

	limit, err := parseLimit(selectStmt)
	if err != nil {
		return DSQLQuery{}, err
	}

	return DSQLQuery{
		Selection:   querySelection,
		Where:       where,
		OrderBy:     orderBy,
		Limit:       limit,
		Fingerprint: generateFingerprint(where),
	}, nil
}

// Function to validate unsupported clauses such as GROUP BY and HAVING
func checkUnsupportedClauses(selectStmt *sqlparser.Select) error {
	if selectStmt.GroupBy != nil || selectStmt.Having != nil {
		return fmt.Errorf("HAVING and GROUP BY clauses are not supported")
	}
	return nil
}

// Function to parse SELECT expressions
func parseSelectExpressions(selectStmt *sqlparser.Select) (QuerySelection, error) {
	if len(selectStmt.SelectExprs) < 1 {
		return QuerySelection{}, fmt.Errorf("no fields selected in result set")
	} else if len(selectStmt.SelectExprs) > 2 {
		return QuerySelection{}, fmt.Errorf("only $key and $value are supported in SELECT expressions")
	}

	keySelection := false
	valueSelection := false
	for _, expr := range selectStmt.SelectExprs {
		aliasedExpr, ok := expr.(*sqlparser.AliasedExpr)
		if !ok {
			return QuerySelection{}, fmt.Errorf("error parsing SELECT expression: %v", expr)
		}
		colName, ok := aliasedExpr.Expr.(*sqlparser.ColName)
		if !ok {
			return QuerySelection{}, fmt.Errorf("only column names are supported in SELECT")
		}
		switch colName.Name.String() {
		case TempKey:
			keySelection = true
		case TempValue:
			valueSelection = true
		default:
			return QuerySelection{}, fmt.Errorf("only $key and $value are supported in SELECT expressions")
		}
	}

	return QuerySelection{KeySelection: keySelection, ValueSelection: valueSelection}, nil
}

// Function to parse table name
func parseTableName(selectStmt *sqlparser.Select) error {
	tableExpr, ok := selectStmt.From[0].(*sqlparser.AliasedTableExpr)
	if !ok {
		return fmt.Errorf("error parsing table name")
	}

	// Remove backticks from table name if present.
	tableName := strings.Trim(sqlparser.String(tableExpr.Expr), "`")

	// Ensure table name is not dual, which means no table name was provided.
	if tableName != "dual" {
		return fmt.Errorf("FROM clause is not supported")
	}

	return nil
}

// Function to parse ORDER BY clause
func parseOrderBy(selectStmt *sqlparser.Select) (QueryOrder, error) {
	orderBy := QueryOrder{}

	// Support only one ORDER BY clause
	if len(selectStmt.OrderBy) > 1 {
		return QueryOrder{}, fmt.Errorf("only one ORDER BY clause is supported")
	}

	if len(selectStmt.OrderBy) == 0 {
		// No ORDER BY clause, return empty order
		return orderBy, nil
	}

	// Extract the ORDER BY expression
	orderExpr := strings.Trim(sqlparser.String(selectStmt.OrderBy[0].Expr), "`")
	orderExpr = trimQuotesOrBackticks(orderExpr)

	// Validate that ORDER BY is either $key or $value
	if orderExpr != TempKey && orderExpr != TempValue && !strings.HasPrefix(orderExpr, TempValue) {
		return QueryOrder{}, fmt.Errorf("only $key and $value expressions are supported in ORDER BY clause")
	}

	// Assign values to QueryOrder
	orderBy.OrderBy = orderExpr
	orderBy.Order = selectStmt.OrderBy[0].Direction

	return orderBy, nil
}

// Helper function to trim both single and double quotes/backticks
func trimQuotesOrBackticks(input string) string {
	if len(input) > 1 && ((input[0] == '\'' && input[len(input)-1] == '\'') ||
		(input[0] == '`' && input[len(input)-1] == '`')) {
		return input[1 : len(input)-1]
	}
	return input
}

// Function to parse LIMIT clause
func parseLimit(selectStmt *sqlparser.Select) (int, error) {
	limit := 0
	if selectStmt.Limit != nil {
		limitVal, err := strconv.Atoi(sqlparser.String(selectStmt.Limit.Rowcount))
		if err != nil {
			return 0, fmt.Errorf("invalid LIMIT value")
		}
		limit = limitVal
	}
	return limit, nil
}

// Function to parse WHERE clause
func parseWhere(selectStmt *sqlparser.Select) sqlparser.Expr {
	if selectStmt.Where == nil {
		return nil
	}
	return selectStmt.Where.Expr
}
