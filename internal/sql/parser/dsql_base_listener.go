// Code generated from internal/sql/parser/dsql.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // dsql

import "github.com/antlr4-go/antlr/v4"

// BasedsqlListener is a complete listener for a parse tree produced by dsqlParser.
type BasedsqlListener struct{}

var _ dsqlListener = &BasedsqlListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BasedsqlListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BasedsqlListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BasedsqlListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BasedsqlListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterQuery is called when production query is entered.
func (s *BasedsqlListener) EnterQuery(ctx *QueryContext) {}

// ExitQuery is called when production query is exited.
func (s *BasedsqlListener) ExitQuery(ctx *QueryContext) {}

// EnterSelectStmt is called when production selectStmt is entered.
func (s *BasedsqlListener) EnterSelectStmt(ctx *SelectStmtContext) {}

// ExitSelectStmt is called when production selectStmt is exited.
func (s *BasedsqlListener) ExitSelectStmt(ctx *SelectStmtContext) {}

// EnterSelectFields is called when production selectFields is entered.
func (s *BasedsqlListener) EnterSelectFields(ctx *SelectFieldsContext) {}

// ExitSelectFields is called when production selectFields is exited.
func (s *BasedsqlListener) ExitSelectFields(ctx *SelectFieldsContext) {}

// EnterWhereClause is called when production whereClause is entered.
func (s *BasedsqlListener) EnterWhereClause(ctx *WhereClauseContext) {}

// ExitWhereClause is called when production whereClause is exited.
func (s *BasedsqlListener) ExitWhereClause(ctx *WhereClauseContext) {}

// EnterOrderByClause is called when production orderByClause is entered.
func (s *BasedsqlListener) EnterOrderByClause(ctx *OrderByClauseContext) {}

// ExitOrderByClause is called when production orderByClause is exited.
func (s *BasedsqlListener) ExitOrderByClause(ctx *OrderByClauseContext) {}

// EnterLimitClause is called when production limitClause is entered.
func (s *BasedsqlListener) EnterLimitClause(ctx *LimitClauseContext) {}

// ExitLimitClause is called when production limitClause is exited.
func (s *BasedsqlListener) ExitLimitClause(ctx *LimitClauseContext) {}

// EnterField is called when production field is entered.
func (s *BasedsqlListener) EnterField(ctx *FieldContext) {}

// ExitField is called when production field is exited.
func (s *BasedsqlListener) ExitField(ctx *FieldContext) {}

// EnterCondition is called when production condition is entered.
func (s *BasedsqlListener) EnterCondition(ctx *ConditionContext) {}

// ExitCondition is called when production condition is exited.
func (s *BasedsqlListener) ExitCondition(ctx *ConditionContext) {}

// EnterOrderByField is called when production orderByField is entered.
func (s *BasedsqlListener) EnterOrderByField(ctx *OrderByFieldContext) {}

// ExitOrderByField is called when production orderByField is exited.
func (s *BasedsqlListener) ExitOrderByField(ctx *OrderByFieldContext) {}

// EnterExpression is called when production expression is entered.
func (s *BasedsqlListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BasedsqlListener) ExitExpression(ctx *ExpressionContext) {}

// EnterFieldWithString is called when production fieldWithString is entered.
func (s *BasedsqlListener) EnterFieldWithString(ctx *FieldWithStringContext) {}

// ExitFieldWithString is called when production fieldWithString is exited.
func (s *BasedsqlListener) ExitFieldWithString(ctx *FieldWithStringContext) {}

// EnterValue is called when production value is entered.
func (s *BasedsqlListener) EnterValue(ctx *ValueContext) {}

// ExitValue is called when production value is exited.
func (s *BasedsqlListener) ExitValue(ctx *ValueContext) {}

// EnterComparisonOp is called when production comparisonOp is entered.
func (s *BasedsqlListener) EnterComparisonOp(ctx *ComparisonOpContext) {}

// ExitComparisonOp is called when production comparisonOp is exited.
func (s *BasedsqlListener) ExitComparisonOp(ctx *ComparisonOpContext) {}
