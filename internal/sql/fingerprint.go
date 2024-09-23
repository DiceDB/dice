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
		return CombineAnd(leftExpr, rightExpr)
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

func CombineAnd(a, b Expression) Expression {
	result := [][]string{}
	for _, termA := range a {
		for _, termB := range b {
			combined := append(termA, termB...)
			uniqueCombined := removeDuplicates(combined)
			result = append(result, uniqueCombined)
		}
	}
	return Expression(result)
}

func combineOr(a, b Expression) Expression {
	result := append(a, b...)
	return Expression(result)
}

// helper
func removeDuplicates(input []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, v := range input {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
