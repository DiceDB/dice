package sql

import (
	"fmt"
	"sort"
	"strings"

	hash "github.com/dgryski/go-farm"
	"github.com/xwb1989/sqlparser"
)

// OR terms containing AND expressions
type expression [][]string

func (expr expression) String() string {
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

func generateFingerprint(where sqlparser.Expr) string {
	expr := parseAstExpression(where)
	return fmt.Sprintf("f_%d", hash.Hash64([]byte(expr.String())))
}

func parseAstExpression(expr sqlparser.Expr) expression {
	switch expr := expr.(type) {
	case *sqlparser.AndExpr:
		leftExpr := parseAstExpression(expr.Left)
		rightExpr := parseAstExpression(expr.Right)
		return combineAnd(leftExpr, rightExpr)
	case *sqlparser.OrExpr:
		leftExpr := parseAstExpression(expr.Left)
		rightExpr := parseAstExpression(expr.Right)
		return combineOr(leftExpr, rightExpr)
	case *sqlparser.ParenExpr:
		return parseAstExpression(expr.Expr)
	case *sqlparser.ComparisonExpr:
		return expression([][]string{{expr.Operator + sqlparser.String(expr.Left) + sqlparser.String(expr.Right)}})
	default:
		return expression{}
	}
}

func combineAnd(a, b expression) expression {
	result := make(expression, 0, len(a)+len(b))
	for _, termA := range a {
		for _, termB := range b {
			combined := make([]string, len(termA), len(termA)+len(termB))
			copy(combined, termA)
			combined = append(combined, termB...)
			uniqueCombined := removeDuplicates(combined)
			sort.Strings(uniqueCombined)
			result = append(result, uniqueCombined)
		}
	}
	return result
}

func combineOr(a, b expression) expression {
	result := make(expression, 0, len(a)+len(b))
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
