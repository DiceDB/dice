// Code generated from DSQL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // DSQL

import "github.com/antlr4-go/antlr/v4"

// DSQLListener is a complete listener for a parse tree produced by DSQLParser.
type DSQLListener interface {
	antlr.ParseTreeListener

	// EnterParse is called when entering the parse production.
	EnterParse(c *ParseContext)

	// EnterSql_stmt_list is called when entering the sql_stmt_list production.
	EnterSql_stmt_list(c *Sql_stmt_listContext)

	// EnterSql_stmt is called when entering the sql_stmt production.
	EnterSql_stmt(c *Sql_stmtContext)

	// EnterSelect_stmt is called when entering the select_stmt production.
	EnterSelect_stmt(c *Select_stmtContext)

	// EnterSelect_core is called when entering the select_core production.
	EnterSelect_core(c *Select_coreContext)

	// EnterFrom_clause is called when entering the from_clause production.
	EnterFrom_clause(c *From_clauseContext)

	// EnterWhere_clause is called when entering the where_clause production.
	EnterWhere_clause(c *Where_clauseContext)

	// EnterOrdering_terms is called when entering the ordering_terms production.
	EnterOrdering_terms(c *Ordering_termsContext)

	// EnterOrdering_term is called when entering the ordering_term production.
	EnterOrdering_term(c *Ordering_termContext)

	// EnterLimit_clause is called when entering the limit_clause production.
	EnterLimit_clause(c *Limit_clauseContext)

	// EnterScalar_expr is called when entering the scalar_expr production.
	EnterScalar_expr(c *Scalar_exprContext)

	// EnterTable_name is called when entering the table_name production.
	EnterTable_name(c *Table_nameContext)

	// EnterLiteral_value is called when entering the literal_value production.
	EnterLiteral_value(c *Literal_valueContext)

	// EnterUnary_operator is called when entering the unary_operator production.
	EnterUnary_operator(c *Unary_operatorContext)

	// EnterBinary_operator is called when entering the binary_operator production.
	EnterBinary_operator(c *Binary_operatorContext)

	// EnterKeyword is called when entering the keyword production.
	EnterKeyword(c *KeywordContext)

	// ExitParse is called when exiting the parse production.
	ExitParse(c *ParseContext)

	// ExitSql_stmt_list is called when exiting the sql_stmt_list production.
	ExitSql_stmt_list(c *Sql_stmt_listContext)

	// ExitSql_stmt is called when exiting the sql_stmt production.
	ExitSql_stmt(c *Sql_stmtContext)

	// ExitSelect_stmt is called when exiting the select_stmt production.
	ExitSelect_stmt(c *Select_stmtContext)

	// ExitSelect_core is called when exiting the select_core production.
	ExitSelect_core(c *Select_coreContext)

	// ExitFrom_clause is called when exiting the from_clause production.
	ExitFrom_clause(c *From_clauseContext)

	// ExitWhere_clause is called when exiting the where_clause production.
	ExitWhere_clause(c *Where_clauseContext)

	// ExitOrdering_terms is called when exiting the ordering_terms production.
	ExitOrdering_terms(c *Ordering_termsContext)

	// ExitOrdering_term is called when exiting the ordering_term production.
	ExitOrdering_term(c *Ordering_termContext)

	// ExitLimit_clause is called when exiting the limit_clause production.
	ExitLimit_clause(c *Limit_clauseContext)

	// ExitScalar_expr is called when exiting the scalar_expr production.
	ExitScalar_expr(c *Scalar_exprContext)

	// ExitTable_name is called when exiting the table_name production.
	ExitTable_name(c *Table_nameContext)

	// ExitLiteral_value is called when exiting the literal_value production.
	ExitLiteral_value(c *Literal_valueContext)

	// ExitUnary_operator is called when exiting the unary_operator production.
	ExitUnary_operator(c *Unary_operatorContext)

	// ExitBinary_operator is called when exiting the binary_operator production.
	ExitBinary_operator(c *Binary_operatorContext)

	// ExitKeyword is called when exiting the keyword production.
	ExitKeyword(c *KeywordContext)
}
