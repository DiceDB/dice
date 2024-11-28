// Code generated from internal/sql/parser/dsql.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // dsql

import "github.com/antlr4-go/antlr/v4"

// dsqlListener is a complete listener for a parse tree produced by dsqlParser.
type dsqlListener interface {
	antlr.ParseTreeListener

	// EnterQuery is called when entering the query production.
	EnterQuery(c *QueryContext)

	// EnterSelectStmt is called when entering the selectStmt production.
	EnterSelectStmt(c *SelectStmtContext)

	// EnterSelectFields is called when entering the selectFields production.
	EnterSelectFields(c *SelectFieldsContext)

	// EnterWhereClause is called when entering the whereClause production.
	EnterWhereClause(c *WhereClauseContext)

	// EnterOrderByClause is called when entering the orderByClause production.
	EnterOrderByClause(c *OrderByClauseContext)

	// EnterLimitClause is called when entering the limitClause production.
	EnterLimitClause(c *LimitClauseContext)

	// EnterField is called when entering the field production.
	EnterField(c *FieldContext)

	// EnterCondition is called when entering the condition production.
	EnterCondition(c *ConditionContext)

	// EnterOrderByField is called when entering the orderByField production.
	EnterOrderByField(c *OrderByFieldContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterFieldWithString is called when entering the fieldWithString production.
	EnterFieldWithString(c *FieldWithStringContext)

	// EnterValue is called when entering the value production.
	EnterValue(c *ValueContext)

	// EnterComparisonOp is called when entering the comparisonOp production.
	EnterComparisonOp(c *ComparisonOpContext)

	// ExitQuery is called when exiting the query production.
	ExitQuery(c *QueryContext)

	// ExitSelectStmt is called when exiting the selectStmt production.
	ExitSelectStmt(c *SelectStmtContext)

	// ExitSelectFields is called when exiting the selectFields production.
	ExitSelectFields(c *SelectFieldsContext)

	// ExitWhereClause is called when exiting the whereClause production.
	ExitWhereClause(c *WhereClauseContext)

	// ExitOrderByClause is called when exiting the orderByClause production.
	ExitOrderByClause(c *OrderByClauseContext)

	// ExitLimitClause is called when exiting the limitClause production.
	ExitLimitClause(c *LimitClauseContext)

	// ExitField is called when exiting the field production.
	ExitField(c *FieldContext)

	// ExitCondition is called when exiting the condition production.
	ExitCondition(c *ConditionContext)

	// ExitOrderByField is called when exiting the orderByField production.
	ExitOrderByField(c *OrderByFieldContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitFieldWithString is called when exiting the fieldWithString production.
	ExitFieldWithString(c *FieldWithStringContext)

	// ExitValue is called when exiting the value production.
	ExitValue(c *ValueContext)

	// ExitComparisonOp is called when exiting the comparisonOp production.
	ExitComparisonOp(c *ComparisonOpContext)
}
