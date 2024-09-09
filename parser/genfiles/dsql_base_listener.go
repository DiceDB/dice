// Code generated from DSQL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // DSQL

import "github.com/antlr4-go/antlr/v4"

// BaseDSQLListener is a complete listener for a parse tree produced by DSQLParser.
type BaseDSQLListener struct{}

var _ DSQLListener = &BaseDSQLListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseDSQLListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseDSQLListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseDSQLListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseDSQLListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterParse is called when production parse is entered.
func (s *BaseDSQLListener) EnterParse(ctx *ParseContext) {}

// ExitParse is called when production parse is exited.
func (s *BaseDSQLListener) ExitParse(ctx *ParseContext) {}

// EnterSql_stmt_list is called when production sql_stmt_list is entered.
func (s *BaseDSQLListener) EnterSql_stmt_list(ctx *Sql_stmt_listContext) {}

// ExitSql_stmt_list is called when production sql_stmt_list is exited.
func (s *BaseDSQLListener) ExitSql_stmt_list(ctx *Sql_stmt_listContext) {}

// EnterSql_stmt is called when production sql_stmt is entered.
func (s *BaseDSQLListener) EnterSql_stmt(ctx *Sql_stmtContext) {}

// ExitSql_stmt is called when production sql_stmt is exited.
func (s *BaseDSQLListener) ExitSql_stmt(ctx *Sql_stmtContext) {}

// EnterSelect_stmt is called when production select_stmt is entered.
func (s *BaseDSQLListener) EnterSelect_stmt(ctx *Select_stmtContext) {}

// ExitSelect_stmt is called when production select_stmt is exited.
func (s *BaseDSQLListener) ExitSelect_stmt(ctx *Select_stmtContext) {}

// EnterSelect_clause is called when production select_clause is entered.
func (s *BaseDSQLListener) EnterSelect_clause(ctx *Select_clauseContext) {}

// ExitSelect_clause is called when production select_clause is exited.
func (s *BaseDSQLListener) ExitSelect_clause(ctx *Select_clauseContext) {}

// EnterFrom_clause is called when production from_clause is entered.
func (s *BaseDSQLListener) EnterFrom_clause(ctx *From_clauseContext) {}

// ExitFrom_clause is called when production from_clause is exited.
func (s *BaseDSQLListener) ExitFrom_clause(ctx *From_clauseContext) {}

// EnterWhere_clause is called when production where_clause is entered.
func (s *BaseDSQLListener) EnterWhere_clause(ctx *Where_clauseContext) {}

// ExitWhere_clause is called when production where_clause is exited.
func (s *BaseDSQLListener) ExitWhere_clause(ctx *Where_clauseContext) {}

// EnterOrder_by_clause is called when production order_by_clause is entered.
func (s *BaseDSQLListener) EnterOrder_by_clause(ctx *Order_by_clauseContext) {}

// ExitOrder_by_clause is called when production order_by_clause is exited.
func (s *BaseDSQLListener) ExitOrder_by_clause(ctx *Order_by_clauseContext) {}

// EnterLimit_clause is called when production limit_clause is entered.
func (s *BaseDSQLListener) EnterLimit_clause(ctx *Limit_clauseContext) {}

// ExitLimit_clause is called when production limit_clause is exited.
func (s *BaseDSQLListener) ExitLimit_clause(ctx *Limit_clauseContext) {}

// EnterResult_column is called when production result_column is entered.
func (s *BaseDSQLListener) EnterResult_column(ctx *Result_columnContext) {}

// ExitResult_column is called when production result_column is exited.
func (s *BaseDSQLListener) ExitResult_column(ctx *Result_columnContext) {}

// EnterTable_name is called when production table_name is entered.
func (s *BaseDSQLListener) EnterTable_name(ctx *Table_nameContext) {}

// ExitTable_name is called when production table_name is exited.
func (s *BaseDSQLListener) ExitTable_name(ctx *Table_nameContext) {}

// EnterExpr is called when production expr is entered.
func (s *BaseDSQLListener) EnterExpr(ctx *ExprContext) {}

// ExitExpr is called when production expr is exited.
func (s *BaseDSQLListener) ExitExpr(ctx *ExprContext) {}

// EnterJson_path is called when production json_path is entered.
func (s *BaseDSQLListener) EnterJson_path(ctx *Json_pathContext) {}

// ExitJson_path is called when production json_path is exited.
func (s *BaseDSQLListener) ExitJson_path(ctx *Json_pathContext) {}

// EnterJson_path_part is called when production json_path_part is entered.
func (s *BaseDSQLListener) EnterJson_path_part(ctx *Json_path_partContext) {}

// ExitJson_path_part is called when production json_path_part is exited.
func (s *BaseDSQLListener) ExitJson_path_part(ctx *Json_path_partContext) {}

// EnterOrdering_term is called when production ordering_term is entered.
func (s *BaseDSQLListener) EnterOrdering_term(ctx *Ordering_termContext) {}

// ExitOrdering_term is called when production ordering_term is exited.
func (s *BaseDSQLListener) ExitOrdering_term(ctx *Ordering_termContext) {}

// EnterUnary_operator is called when production unary_operator is entered.
func (s *BaseDSQLListener) EnterUnary_operator(ctx *Unary_operatorContext) {}

// ExitUnary_operator is called when production unary_operator is exited.
func (s *BaseDSQLListener) ExitUnary_operator(ctx *Unary_operatorContext) {}

// EnterBinary_operator is called when production binary_operator is entered.
func (s *BaseDSQLListener) EnterBinary_operator(ctx *Binary_operatorContext) {}

// ExitBinary_operator is called when production binary_operator is exited.
func (s *BaseDSQLListener) ExitBinary_operator(ctx *Binary_operatorContext) {}

// EnterFunction_name is called when production function_name is entered.
func (s *BaseDSQLListener) EnterFunction_name(ctx *Function_nameContext) {}

// ExitFunction_name is called when production function_name is exited.
func (s *BaseDSQLListener) ExitFunction_name(ctx *Function_nameContext) {}

// EnterLiteral_value is called when production literal_value is entered.
func (s *BaseDSQLListener) EnterLiteral_value(ctx *Literal_valueContext) {}

// ExitLiteral_value is called when production literal_value is exited.
func (s *BaseDSQLListener) ExitLiteral_value(ctx *Literal_valueContext) {}

// EnterError is called when production error is entered.
func (s *BaseDSQLListener) EnterError(ctx *ErrorContext) {}

// ExitError is called when production error is exited.
func (s *BaseDSQLListener) ExitError(ctx *ErrorContext) {}
