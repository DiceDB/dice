package sql

import (
	"fmt"
	"sort"
	"strings"

	hash "github.com/dgryski/go-farm"
	"github.com/xwb1989/sqlparser"
)

// OR terms containing AND expressions
type Expression [][]string

func (expr Expression) String() string {
	var orTerms []string
	for _, andTerm := range expr {
		// Sort AND terms within OR
		sort.Strings(andTerm)
		orTerms = append(orTerms, strings.Join(andTerm, " AND "))
	}
	// Sort the OR terms
	sort.Strings(orTerms)
	return strings.Join(orTerms, " OR ")
}

func GenerateFingerprint(where sqlparser.Expr) string {
	expr := ParseAstExpression(where)
	return fmt.Sprintf("f_%d", hash.Hash64([]byte(expr.String())))
}

func ParseAstExpression(expr sqlparser.Expr) Expression {
	switch expr := expr.(type) {
	case *sqlparser.AndExpr:
		leftExpr := ParseAstExpression(expr.Left)
		rightExpr := ParseAstExpression(expr.Right)
		return combineAnd(leftExpr, rightExpr)
	case *sqlparser.OrExpr:
		leftExpr := ParseAstExpression(expr.Left)
		rightExpr := ParseAstExpression(expr.Right)
		return combineOr(leftExpr, rightExpr)
	case *sqlparser.ParenExpr:
		return ParseAstExpression(expr.Expr)
	case *sqlparser.ComparisonExpr:
		return Expression([][]string{{expr.Operator + sqlparser.String(expr.Left) + sqlparser.String(expr.Right)}})
	default:
		return Expression{}
	}
}

func combineAnd(a, b Expression) Expression {
	var result [][]string
	for _, termA := range a {
		for _, termB := range b {
			result = append(result, append(termA, termB...))
		}
	}
	return Expression(result)
}

func combineOr(a, b Expression) Expression {
	result := append(a, b...)
	return Expression(result)
}
