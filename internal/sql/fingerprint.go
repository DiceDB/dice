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
		return CombineOr(leftExpr, rightExpr)
	case *sqlparser.ParenExpr:
		return ParseAstExpression(expr.Expr)
	case *sqlparser.ComparisonExpr:
		return Expression([][]string{{expr.Operator + sqlparser.String(expr.Left) + sqlparser.String(expr.Right)}})
	default:
		return Expression{}
	}
}

func CombineAnd(a, b Expression) Expression {
	result := make(Expression, 0, len(a)+len(b))
	for _, termA := range a {
		for _, termB := range b {
			combined := append(termA, termB...)
			uniqueCombined := removeDuplicates(combined)
			sort.Strings(uniqueCombined)
			result = append(result, uniqueCombined)
		}
	}
	return Expression(result)
}

func CombineOr(a, b Expression) Expression {
	result := make(Expression, 0, len(a)+len(b))
	uniqueTerms := make(map[string]bool)

	// Helper function to add unique terms
	addUnique := func(terms []string) {
		// Sort the terms for consistent ordering
		sort.Strings(terms)
		key := strings.Join(terms, ",")
		if !uniqueTerms[key] {
			result = append(result, terms)
			uniqueTerms[key] = true
		}
	}

	// Add unique terms from a
	for _, terms := range a {
		addUnique(append([]string(nil), terms...))
	}

	// Add unique terms from b
	for _, terms := range b {
		addUnique(append([]string(nil), terms...))
	}

	return result
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
