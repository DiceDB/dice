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

// EnterSelect_core is called when production select_core is entered.
func (s *BaseDSQLListener) EnterSelect_core(ctx *Select_coreContext) {}

// ExitSelect_core is called when production select_core is exited.
func (s *BaseDSQLListener) ExitSelect_core(ctx *Select_coreContext) {}

// EnterFrom_clause is called when production from_clause is entered.
func (s *BaseDSQLListener) EnterFrom_clause(ctx *From_clauseContext) {}

// ExitFrom_clause is called when production from_clause is exited.
func (s *BaseDSQLListener) ExitFrom_clause(ctx *From_clauseContext) {}

// EnterWhere_clause is called when production where_clause is entered.
func (s *BaseDSQLListener) EnterWhere_clause(ctx *Where_clauseContext) {}

// ExitWhere_clause is called when production where_clause is exited.
func (s *BaseDSQLListener) ExitWhere_clause(ctx *Where_clauseContext) {}

// EnterOrdering_terms is called when production ordering_terms is entered.
func (s *BaseDSQLListener) EnterOrdering_terms(ctx *Ordering_termsContext) {}

// ExitOrdering_terms is called when production ordering_terms is exited.
func (s *BaseDSQLListener) ExitOrdering_terms(ctx *Ordering_termsContext) {}

// EnterOrdering_term is called when production ordering_term is entered.
func (s *BaseDSQLListener) EnterOrdering_term(ctx *Ordering_termContext) {}

// ExitOrdering_term is called when production ordering_term is exited.
func (s *BaseDSQLListener) ExitOrdering_term(ctx *Ordering_termContext) {}

// EnterLimit_clause is called when production limit_clause is entered.
func (s *BaseDSQLListener) EnterLimit_clause(ctx *Limit_clauseContext) {}

// ExitLimit_clause is called when production limit_clause is exited.
func (s *BaseDSQLListener) ExitLimit_clause(ctx *Limit_clauseContext) {}

// EnterScalar_expr is called when production scalar_expr is entered.
func (s *BaseDSQLListener) EnterScalar_expr(ctx *Scalar_exprContext) {}

// ExitScalar_expr is called when production scalar_expr is exited.
func (s *BaseDSQLListener) ExitScalar_expr(ctx *Scalar_exprContext) {}

// EnterTable_name is called when production table_name is entered.
func (s *BaseDSQLListener) EnterTable_name(ctx *Table_nameContext) {}

// ExitTable_name is called when production table_name is exited.
func (s *BaseDSQLListener) ExitTable_name(ctx *Table_nameContext) {}

// EnterLiteral_value is called when production literal_value is entered.
func (s *BaseDSQLListener) EnterLiteral_value(ctx *Literal_valueContext) {}

// ExitLiteral_value is called when production literal_value is exited.
func (s *BaseDSQLListener) ExitLiteral_value(ctx *Literal_valueContext) {}

// EnterUnary_operator is called when production unary_operator is entered.
func (s *BaseDSQLListener) EnterUnary_operator(ctx *Unary_operatorContext) {}

// ExitUnary_operator is called when production unary_operator is exited.
func (s *BaseDSQLListener) ExitUnary_operator(ctx *Unary_operatorContext) {}

// EnterBinary_operator is called when production binary_operator is entered.
func (s *BaseDSQLListener) EnterBinary_operator(ctx *Binary_operatorContext) {}

// ExitBinary_operator is called when production binary_operator is exited.
func (s *BaseDSQLListener) ExitBinary_operator(ctx *Binary_operatorContext) {}

// EnterKeyword is called when production keyword is entered.
func (s *BaseDSQLListener) EnterKeyword(ctx *KeywordContext) {}

// ExitKeyword is called when production keyword is exited.
func (s *BaseDSQLListener) ExitKeyword(ctx *KeywordContext) {}
