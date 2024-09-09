package parser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	parser "github.com/dicedb/dice/parser/genfiles"
	"strconv"
	"strings"
)

// DSQLQuery structure (modified to match your existing structure)
type DSQLQuery struct {
	Selection QuerySelection
	KeyRegex  string
	Where     parser.IExprContext
	OrderBy   []QueryOrder
	Limit     int
}

type QuerySelection struct {
	KeySelection   bool
	ValueSelection bool
}

type QueryOrder struct {
	OrderBy parser.IExprContext
	Order   string
}

// CustomListener to implement the ANTLR listener methods
type CustomListener struct {
	*parser.BaseDSQLListener
	Query DSQLQuery
}
type CustomErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []string
}

func NewCustomErrorListener() *CustomErrorListener {
	return &CustomErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		Errors:               []string{},
	}
}

func (l *CustomErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	l.Errors = append(l.Errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

func Parse(sql string) (query DSQLQuery, err error) {
	// Create a new input stream for the sql
	input := antlr.NewInputStream(sql)

	// Create a lexer that feeds off the input stream
	lexer := parser.NewDSQLLexer(input)

	// Create a token stream from the lexer
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create a parser that feeds off the token stream
	sqlParser := parser.NewDSQLParser(tokens)

	// Create and add the custom error listener
	errorListener := NewCustomErrorListener()
	sqlParser.RemoveErrorListeners() // Remove default error listeners
	sqlParser.AddErrorListener(errorListener)

	// Set up a listener
	listener := &CustomListener{}

	// Parse the SQL and walk the parse tree with the listener
	tree := sqlParser.Parse()

	// Walk the tree
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	// Check for errors
	if len(errorListener.Errors) > 0 {
		fmt.Println("Parsing errors:")
		for _, err := range errorListener.Errors {
			fmt.Println(err)
		}
		return DSQLQuery{}, fmt.Errorf(strings.Join(errorListener.Errors, "\n"))
	}

	return listener.Query, nil
}

func (l *CustomListener) EnterSelect_core(ctx *parser.Select_stmtContext) {
	result := ctx.Select_clause().Result_column()
	if result.STAR() != nil {
		l.Query.Selection.KeySelection = true
		l.Query.Selection.ValueSelection = true
	} else if result.KEY() != nil {
		l.Query.Selection.KeySelection = true
	} else if result.VALUE() != nil {
		l.Query.Selection.ValueSelection = true
	}
}

func (l *CustomListener) EnterFrom_clause(ctx *parser.From_clauseContext) {
	if tableName := ctx.Table_name().GetText(); tableName != "" {
		l.Query.KeyRegex = tableName
	}
}

func (l *CustomListener) EnterWhere_clause(ctx *parser.Where_clauseContext) {
	if expr := ctx.Expr(); expr != nil {
		l.Query.Where = expr
	}
}

func (l *CustomListener) EnterOrdering_term(ctx *parser.Order_by_clauseContext) {
	for _, orderingTerm := range ctx.AllOrdering_term() {
		queryOrder := QueryOrder{}
		if expr := orderingTerm.Expr(); expr != nil {
			queryOrder.OrderBy = expr
		}
		queryOrder.Order = "ASC"
		if orderingTerm.K_ASC() != nil {
			queryOrder.Order = "ASC"
		} else if orderingTerm.K_DESC() != nil {
			queryOrder.Order = "DESC"
		}
		l.Query.OrderBy = append(l.Query.OrderBy, queryOrder)
	}
}

func (l *CustomListener) EnterLimit_clause(ctx *parser.Limit_clauseContext) {
	if limitStr := ctx.NUMERIC_LITERAL().GetText(); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			l.Query.Limit = limit
		}
	}
}
