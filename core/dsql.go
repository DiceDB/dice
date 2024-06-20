// This file contains the core logic for parsing DSQL queries. DSQL stands for
// Dice SQL, and is a simple SQL-like language that is used to query data from
// DiceDB.

package core

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

type UnsupportedDSQLStatementError struct {
	Stmt sqlparser.Statement
}

func (e *UnsupportedDSQLStatementError) Error() string {
	return fmt.Sprintf("unsupported DSQL statement: %T", e.Stmt)
}

func newUnsupportedSqlStatementError(stmt sqlparser.Statement) *UnsupportedDSQLStatementError {
	return &UnsupportedDSQLStatementError{Stmt: stmt}
}

// ParseQuery takes a DSQL string and checks if it is a simple "SELECT <field_name>" statement.
func ParseQuery(sql string) (string, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", fmt.Errorf("error parsing SQL statement: %v", err)
	}

	// Check if the statement is a Select.
	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		return "", newUnsupportedSqlStatementError(stmt)
	}

	// We expect exactly one expression in the SELECT expression list.
	if len(selectStmt.SelectExprs) != 1 {
		return "", fmt.Errorf("only single field selections are supported, found %d fields", len(selectStmt.SelectExprs))
	}

	// Check that the expression is a simple identifier, not a wildcard or something else.
	expr, ok := selectStmt.SelectExprs[0].(*sqlparser.AliasedExpr)
	if !ok {
		return "", fmt.Errorf("only simple field selections are supported")
	}

	// Check that the expr is a ColName.
	colName, ok := expr.Expr.(*sqlparser.ColName)
	if !ok {
		return "", fmt.Errorf("only column names are supported in SELECT")
	}

	return colName.Name.String(), nil
}
