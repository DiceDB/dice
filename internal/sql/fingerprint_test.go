package sql

import (
	"testing"

	"github.com/xwb1989/sqlparser"
	"github.com/stretchr/testify/assert"
)

func TestExpressionString(t *testing.T) {
	tests := []struct {
		name     string
		expr     expression
		expected string
	}{
		{
			name:     "Single AND term",
			expr:     expression{{"_key LIKE 'match:1:*'", "_value > 10"}},
			expected: "_key LIKE 'match:1:*' AND _value > 10",
		},
		{
			name:     "Multiple AND terms in a single OR term",
			expr:     expression{{"_key LIKE 'match:1:*'", "_value > 10", "_value < 5"}},
			expected: "_key LIKE 'match:1:*' AND _value < 5 AND _value > 10",
		},
		{
			name:     "Multiple OR terms with single AND terms",
			expr:     expression{{"_key LIKE 'match:1:*'"}, {"_value < 5"}, {"_value > 10"}},
			expected: "_key LIKE 'match:1:*' OR _value < 5 OR _value > 10",
		},
		{
			name:     "Multiple OR terms with AND combinations",
			expr:     expression{{"_key LIKE 'match:1:*'", "_value > 10"}, {"_value < 5", "_value > 0"}},
			expected: "_key LIKE 'match:1:*' AND _value > 10 OR _value < 5 AND _value > 0",
		},
		{
			name:     "Unordered terms",
			expr:     expression{{"_value > 10", "_key LIKE 'match:1:*'"}, {"_value > 0", "_value < 5"}},
			expected: "_key LIKE 'match:1:*' AND _value > 10 OR _value < 5 AND _value > 0",
		},
		{
			name:     "Nested AND and OR terms with duplicates",
			expr:     expression{{"_key LIKE 'match:1:*'", "_value < 5"}, {"_key LIKE 'match:1:*'", "_value < 5", "_value > 10"}},
			expected: "_key LIKE 'match:1:*' AND _value < 5 OR _key LIKE 'match:1:*' AND _value < 5 AND _value > 10",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.expr.String())
		})
	}
}

func TestCombineOr(t *testing.T) {
	tests := []struct {
		name     string
		a        expression
		b        expression
		expected expression
	}{
		{
			name:     "Combining two empty expressions",
			a:        expression([][]string{}),
			b:        expression([][]string{}),
			expected: expression([][]string{}),
		},
		{
			name:     "Identity law",
			a:        expression([][]string{{"_value > 10"}}),
			b:        expression([][]string{}), // equivalent to 0
			expected: expression([][]string{{"_value > 10"}}),
		},
		{
			name:     "Idempotent law",
			a:        expression([][]string{{"_value > 10"}}),
			b:        expression([][]string{{"_value > 10"}}),
			expected: expression([][]string{{"_value > 10"}}),
		},
		{
			name: "Simple OR combination with non-overlapping terms",
			a:    expression([][]string{{"_key LIKE 'test:*'"}}),
			b:    expression([][]string{{"_value > 10"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'"}, {"_value > 10"},
			}),
		},
		{
			name: "Complex OR combination with multiple AND terms",
			a:    expression([][]string{{"_key LIKE 'test:*'", "_value > 10"}}),
			b:    expression([][]string{{"_key LIKE 'example:*'", "_value < 5"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'", "_value > 10"}, {"_key LIKE 'example:*'", "_value < 5"},
			}),
		},
		{
			name: "Combining overlapping AND terms",
			a:    expression([][]string{{"_key LIKE 'test:*'", "_value > 10"}}),
			b:    expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'", "_value > 10"},
			}),
		},
		{
			name: "Combining overlapping AND terms in reverse order",
			a:    expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
			b:    expression([][]string{{"_key LIKE 'test:*'", "_value > 10"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'", "_value > 10"},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, combineOr(tt.a, tt.b))
		})
	}
}

func TestCombineAnd(t *testing.T) {
	tests := []struct {
		name     string
		a        expression
		b        expression
		expected expression
	}{
		{
			name:     "Combining two empty expressions",
			a:        expression([][]string{}),
			b:        expression([][]string{}),
			expected: expression([][]string{}),
		},
		{
			name:     "Annulment law",
			a:        expression([][]string{{"_value > 10"}}),
			b:        expression([][]string{}), // equivalent to 0
			expected: expression([][]string{}),
		},
		{
			name:     "Identity law",
			a:        expression([][]string{{"_value > 10"}}),
			b:        expression([][]string{{}}), // equivalent to 1
			expected: expression([][]string{{"_value > 10"}}),
		},
		{
			name:     "Idempotent law",
			a:        expression([][]string{{"_value > 10"}}),
			b:        expression([][]string{{"_value > 10"}}),
			expected: expression([][]string{{"_value > 10"}}),
		},
		{
			name:     "Multiple AND terms, no duplicates",
			a:        expression([][]string{{"_value > 10"}}),
			b:        expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}}),
			expected: expression([][]string{{"_key LIKE 'test:*'", "_value < 5", "_value > 10"}}),
		},
		{
			name: "Multiple terms in both expressions with duplicates",
			a:    expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
			b:    expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'", "_value < 5", "_value > 10"},
			}),
		},
		{
			name: "Terms in different order, no duplicates",
			a:    expression([][]string{{"_key LIKE 'test:*'", "_value > 10"}}),
			b:    expression([][]string{{"_value < 5"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'", "_value < 5", "_value > 10"},
			}),
		},
		{
			name: "Terms in different order with duplicates",
			a:    expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
			b:    expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'", "_value < 5", "_value > 10"},
			}),
		},
		{
			name: "Partial duplicates across expressions",
			a:    expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
			b:    expression([][]string{{"_key LIKE 'test:*'", "_key = 'abc'"}}),
			expected: expression([][]string{
				{"_key = 'abc'", "_key LIKE 'test:*'", "_value > 10"},
			}),
		},
		{
			name: "Nested AND groups",
			a:    expression([][]string{{"_key LIKE 'test:*'", "_value > 10"}}),
			b:    expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}, {"_value = 7"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'", "_value < 5", "_value > 10"},
				{"_key LIKE 'test:*'", "_value = 7", "_value > 10"},
			}),
		},
		{
			name: "Same terms but in different AND groups",
			a:    expression([][]string{{"_key LIKE 'test:*'"}}),
			b:    expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}, {"_key LIKE 'test:*'"}}),
			expected: expression([][]string{
				{"_key LIKE 'test:*'", "_value < 5"},
				{"_key LIKE 'test:*'"},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, combineAnd(tt.a, tt.b))
		})
	}
}

func TestGenerateFingerprintAndParseAstExpression(t *testing.T) {
	tests := []struct {
		name        string
		similarExpr []string // logically same where expressions
		expression  string
		fingerprint string
	}{
		{
			name: "Terms in different order, OR operator",
			similarExpr: []string{
				"_value > 10 OR _value < 5",
				"_value < 5 OR _value > 10",
			},
			expression:  "<_value5 OR >_value10",
			fingerprint: "f_5731466836575684070",
		},
		{
			name: "Terms in different order, AND operator",
			similarExpr: []string{
				"_value > 10 AND _value < 5",
				"_value < 5 AND _value > 10",
			},
			expression:  "<_value5 AND >_value10",
			fingerprint: "f_8696580727138087340",
		},
		{
			// ideally this and below test should spit same output
			name: "Simple comparison operator (comparison value in backticks)",
			similarExpr: []string{
				"_key like `match:1:*`",
			},
			expression:  "like_key`match:1:*`",
			fingerprint: "f_15929225480754059748",
		},
		{
			name: "Simple comparison operator (comparison value in single quotes)",
			similarExpr: []string{
				"_key like 'match:1:*'",
			},
			expression:  "like_key'match:1:*'",
			fingerprint: "f_5313097907453016110",
		},
		{
			name: "Simple comparison operator with multiple redundant parentheses",
			similarExpr: []string{
				"_key like 'match:1:*'",
				"(_key like 'match:1:*')",
				"((_key like 'match:1:*'))",
				"(((_key like 'match:1:*')))",
			},
			expression:  "like_key'match:1:*'",
			fingerprint: "f_5313097907453016110",
		},
		{
			name: "Expression with duplicate terms (or Idempotent law)",
			similarExpr: []string{
				"_key like 'match:1:*' AND _key like 'match:1:*'",
				"_key like 'match:1:*'",
			},
			expression:  "like_key'match:1:*'",
			fingerprint: "f_5313097907453016110",
		},
		{
			name: "expression with exactly 1 term, multiple AND OR (Idempotent law)",
			similarExpr: []string{
				"_value > 10 AND _value > 10 OR _value > 10 AND _value > 10",
				"_value > 10",
			},
			expression:  ">_value10",
			fingerprint: "f_11845737393789912467",
		},
		{
			name: "Expression in form 'A AND (B OR C)' which can reduce to 'A AND B OR A AND C' etc (or Distributive law)",
			similarExpr: []string{
				"(_key LIKE 'test:*') AND (_value > 10 OR _value < 5)",
				"(_value > 10 OR _value < 5) AND (_key LIKE 'test:*')",
				"(_value < 5 OR _value > 10) AND (_key LIKE 'test:*')",
				"(_key LIKE 'test:*' AND _value > 10) OR (_key LIKE 'test:*' AND _value < 5)",
				"((_key LIKE 'test:*') AND _value > 10) OR ((_key LIKE 'test:*') AND _value < 5)",
				"(_key LIKE 'test:*') AND ((_value > 10) OR (_value < 5))",
				"(_value > 10 AND _key LIKE 'test:*') OR (_value < 5 AND _key LIKE 'test:*')",
				"(_value < 5 AND _key LIKE 'test:*') OR (_value > 10 AND _key LIKE 'test:*')",
			},
			expression:  "<_value5 AND like_key'test:*' OR >_value10 AND like_key'test:*'",
			fingerprint: "f_6936111135456499050",
		},
		{
			// ideally this and below test should spit same output
			// but our algorithm is not sophisticated enough yet
			name: "Expression in form 'A OR (B AND C)' which can reduce to 'A OR B AND A OR C' etc (or Distributive law)",
			similarExpr: []string{
				"_key LIKE 'test:*' OR _value > 10 AND _value < 5",
				"(_key LIKE 'test:*') OR (_value > 10 AND _value < 5)",
				"(_value > 10 AND _value < 5) OR (_key LIKE 'test:*')",
				"(_value < 5 AND _value > 10) OR (_key LIKE 'test:*')",
				// "(_key LIKE 'test:*' OR _value > 10) AND (_key LIKE 'test:*' OR _value < 5)",
				// "((_key LIKE 'test:*') OR (_value > 10)) AND ((_key LIKE 'test:*') OR (_value < 5))",
			},
			expression:  "<_value5 AND >_value10 OR like_key'test:*'",
			fingerprint: "f_655732287561200780",
		},
		{
			name: "Complex expression with multiple redundant parentheses",
			similarExpr: []string{
				"(_key LIKE 'test:*' OR _value > 10) AND (_key LIKE 'test:*' OR _value < 5)",
				"((_key LIKE 'test:*') OR (_value > 10)) AND ((_key LIKE 'test:*') OR (_value < 5))",
			},
			expression:  "<_value5 AND >_value10 OR <_value5 AND like_key'test:*' OR >_value10 AND like_key'test:*' OR like_key'test:*'",
			fingerprint: "f_1509117529358989129",
		},
		{
			name: "Test Precedence: AND before OR with LIKE and Value Comparison",
			similarExpr: []string{
				"_key LIKE 'test:*' AND _value > 10 OR _value < 5",
				"(_key LIKE 'test:*' AND _value > 10) OR _value < 5",
			},
			expression:  "<_value5 OR >_value10 AND like_key'test:*'",
			fingerprint: "f_8791273852316817684",
		},
		{
			name: "Simple JSON expression",
			similarExpr: []string{
				"'_value.age' > 30 and _key like 'user:*' and '_value.age' > 30",
				"'_value.age' > 30 and _key like 'user:*'",
				"_key like 'user:*' and '_value.age' > 30 ",
			},
			expression:  ">'_value.age'30 AND like_key'user:*'",
			fingerprint: "f_5016002712062179335",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, query := range tt.similarExpr {
				where, err := parseSQL("SELECT * WHERE " + query)
				if err != nil {
					t.Fail()
				}
				assert.Equal(t, tt.expression, parseAstExpression(where).String())
				assert.Equal(t, tt.fingerprint, generateFingerprint(where))
			}
		})
	}
}

// Benchmark for generateFingerprint function
func BenchmarkGenerateFingerprint(b *testing.B) {
	queries := []struct {
		name  string
		query string
	}{
		{"Simple", "SELECT * WHERE _key LIKE 'match:1:*'"},
		{"OrExpression", "SELECT * WHERE _key LIKE 'match:1:*' OR _value > 10"},
		{"AndExpression", "SELECT * WHERE _key LIKE 'match:1:*' AND _value > 10"},
		{"NestedOrAnd", "SELECT * WHERE _key LIKE 'match:1:*' OR (_value > 10 AND _value < 5)"},
		{"DeepNested", "SELECT * FROM table WHERE _key LIKE 'match:1:*' OR (_value > 10 AND (_value < 5 OR '_value.age' > 18))"},
	}

	for _, tt := range queries {
		expr, err := parseSQL(tt.query)
		if err != nil {
			b.Fail()
		}

		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				generateFingerprint(expr)
			}
		})
	}
}

// helper
func parseSQL(query string) (sqlparser.Expr, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, err
	}

	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		return nil, err
	}

	return selectStmt.Where.Expr, nil
}
