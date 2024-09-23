package sql_test

import (
	"testing"

	"github.com/dicedb/dice/internal/sql"
	"github.com/xwb1989/sqlparser"
	"gotest.tools/v3/assert"
)

func TestCombineAnd(t *testing.T) {
	tests := []struct {
		name     string
		a        sql.Expression
		b        sql.Expression
		expected sql.Expression
	}{
		{
			name:     "Simple combine with no duplicates",
			a:        sql.Expression([][]string{{"_value > 10"}}),
			b:        sql.Expression([][]string{{"_key LIKE 'test:*'"}}),
			expected: sql.Expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
		},
		{
			name:     "Combine with duplicate values",
			a:        sql.Expression([][]string{{"_value > 10"}}),
			b:        sql.Expression([][]string{{"_value > 10"}}),
			expected: sql.Expression([][]string{{"_value > 10"}}),
		},
		{
			name:     "Multiple AND terms, no duplicates",
			a:        sql.Expression([][]string{{"_value > 10"}}),
			b:        sql.Expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}}),
			expected: sql.Expression([][]string{{"_value > 10", "_key LIKE 'test:*'", "_value < 5"}}),
		},
		{
			name: "Multiple terms in both expressions with duplicates",
			a:    sql.Expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
			b:    sql.Expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}}),
			expected: sql.Expression([][]string{
				{"_value > 10", "_key LIKE 'test:*'", "_value < 5"},
			}),
		},
		{
			name: "Terms in different order, no duplicates",
			a:    sql.Expression([][]string{{"_key LIKE 'test:*'", "_value > 10"}}),
			b:    sql.Expression([][]string{{"_value < 5"}}),
			expected: sql.Expression([][]string{
				{"_key LIKE 'test:*'", "_value > 10", "_value < 5"},
			}),
		},
		{
			name: "Terms in different order with duplicates",
			a:    sql.Expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
			b:    sql.Expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}}),
			expected: sql.Expression([][]string{
				{"_value > 10", "_key LIKE 'test:*'", "_value < 5"},
			}),
		},
		{
			name: "Partial duplicates across expressions",
			a:    sql.Expression([][]string{{"_value > 10", "_key LIKE 'test:*'"}}),
			b:    sql.Expression([][]string{{"_key LIKE 'test:*'", "_key = 'abc'"}}),
			expected: sql.Expression([][]string{
				{"_value > 10", "_key LIKE 'test:*'", "_key = 'abc'"},
			}),
		},
		{
			name: "Nested AND groups",
			a:    sql.Expression([][]string{{"_key LIKE 'test:*'", "_value > 10"}}),
			b:    sql.Expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}, {"_value = 7"}}),
			expected: sql.Expression([][]string{
				{"_key LIKE 'test:*'", "_value > 10", "_value < 5"},
				{"_key LIKE 'test:*'", "_value > 10", "_value = 7"},
			}),
		},
		{
			name: "Same terms but in different AND groups",
			a:    sql.Expression([][]string{{"_key LIKE 'test:*'"}}),
			b:    sql.Expression([][]string{{"_key LIKE 'test:*'", "_value < 5"}, {"_key LIKE 'test:*'"}}),
			expected: sql.Expression([][]string{
				{"_key LIKE 'test:*'", "_value < 5"},
				{"_key LIKE 'test:*'"},
			}),
		},
		{
			name:     "Empty expression",
			a:        sql.Expression([][]string{}),
			b:        sql.Expression([][]string{}),
			expected: sql.Expression([][]string{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, tt.expected, sql.CombineAnd(tt.a, tt.b))
		})
	}
}

func TestParseAstExpression(t *testing.T) {
	tests := []struct {
		name       string
		similarEq  []string // logically same where expressions
		expression string
	}{
		{
			name: "aa",
			similarEq: []string{
				"_value > 10 OR _value < 5",
				" _value < 5 OR _value > 10",
			},
			expression: "<_value5 OR >_value10",
		},
		{
			name: "aa",
			similarEq: []string{
				"_value > 10 AND _value < 5",
				"_value < 5 AND _value > 10",
			},
			expression: "<_value5 AND >_value10",
		},
		{
			name: "aa",
			similarEq: []string{
				"_key like `match:1:*`",
			},
			expression: "like_key`match:1:*`",
		},
		{
			name: "aa",
			similarEq: []string{
				"_key like 'match:1:*'",
			},
			expression: "like_key'match:1:*'",
		},
		{
			name: "aa",
			similarEq: []string{
				"_key like 'match:1:*'",
				"(_key like 'match:1:*')",
				"((_key like 'match:1:*'))",
				"(((_key like 'match:1:*')))",
			},
			expression: "like_key'match:1:*'",
		},
		{
			name: "pskd",
			similarEq: []string{
				"_key like 'match:1:*' AND _key like 'match:1:*'",
				"_key like 'match:1:*'",
			},
			expression: "like_key'match:1:*'",
		},
		{
			name: "dcdjfhcdipp",
			similarEq: []string{
				"(_key LIKE 'test:*') AND (_value > 10 OR _value < 5)",
				"(_value > 10 OR _value < 5) AND (_key LIKE 'test:*')",
				"(_value < 5 OR _value > 10) AND (_key LIKE 'test:*')",
				"(_key LIKE 'test:*' AND _value > 10) OR (_key LIKE 'test:*' AND _value < 5)",
				"((_key LIKE 'test:*') AND _value > 10) OR ((_key LIKE 'test:*') AND _value < 5)",
				"(_key LIKE 'test:*') AND ((_value > 10) OR (_value < 5))",
				"(_value > 10 AND _key LIKE 'test:*') OR (_value < 5 AND _key LIKE 'test:*')",
				"(_value < 5 AND _key LIKE 'test:*') OR (_value > 10 AND _key LIKE 'test:*')",
			},
			expression: "<_value5 AND like_key'test:*' OR >_value10 AND like_key'test:*'",
		},
		// {
		// 	name: "fjvh",
		// 	similarEq: []string{
		// 		// "(_key LIKE 'test:*') OR (_value > 10 AND _value < 5)", // "<_value5 AND >_value10 OR like_key'test:*'"
		// 		// "(_value > 10 AND _value < 5) OR (_key LIKE 'test:*')", // "<_value5 AND >_value10 OR like_key'test:*'"
		// 		// "(_value < 5 AND _value > 10) OR (_key LIKE 'test:*')", // "<_value5 AND >_value10 OR like_key'test:*'"
		// 		// "(_key LIKE 'test:*' OR _value > 10) AND (_key LIKE 'test:*' OR _value < 5)",
		// 		// {
		// 		// 	  	"<_value5 AND >_value10 OR",
		// 		// 	+ 	" <_value5 AND like_key'test:*' OR >_value10 AND like_key'test:*'",
		// 		// 	+ 	" OR like_key'test:*' AND",
		// 		// 	  	" like_key'test:*'",
		// 		// 	  }

		// 		// "((_key LIKE 'test:*') OR (_value > 10)) AND ((_key LIKE 'test:*') OR (_value < 5))",
		// 	},
		// 	expression: "<_value5 AND >_value10 OR like_key'test:*'",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, query := range tt.similarEq {
				statement, err := sqlparser.Parse("SELECT * WHERE " + query)
				if err != nil {
					t.Fail()
				}
				selectStmt, ok := statement.(*sqlparser.Select)
				if !ok {
					t.Fail()
				}
				assert.DeepEqual(t, tt.expression, sql.ParseAstExpression(selectStmt.Where.Expr).String())
			}
		})
	}
}

func TestGenerateFingerprint(t *testing.T) {
	tests := []struct {
		name        string
		similarEq   []string // logically same where expressions
		fingerprint string   // expected fingerprint
	}{
		{
			name: "aa",
			similarEq: []string{
				"_value > 10 OR _value < 5",
				" _value < 5 OR _value > 10",
			},
			fingerprint: "f_5731466836575684070",
		},
		{
			name: "aa",
			similarEq: []string{
				"_value > 10 AND _value < 5",
				"_value < 5 AND _value > 10",
			},
			fingerprint: "f_8696580727138087340",
		},
		{
			name: "aa",
			similarEq: []string{
				"_key like `match:1:*`",
			},
			fingerprint: "f_15929225480754059748",
		},
		{
			name: "aa",
			similarEq: []string{
				"_key like 'match:1:*'",
			},
			fingerprint: "f_5313097907453016110",
		},
		{
			name: "aa",
			similarEq: []string{
				"_key like 'match:1:*'",
				"(_key like 'match:1:*')",
				"((_key like 'match:1:*'))",
				"(((_key like 'match:1:*')))",
			},
			fingerprint: "f_5313097907453016110",
		},
		{
			name: "dcdjfhcdipp",
			similarEq: []string{
				"(_key LIKE 'test:*') AND (_value > 10 OR _value < 5)",
				"(_value > 10 OR _value < 5) AND (_key LIKE 'test:*')",
				"(_value < 5 OR _value > 10) AND (_key LIKE 'test:*')",
				"(_key LIKE 'test:*' AND _value > 10) OR (_key LIKE 'test:*' AND _value < 5)",
				"((_key LIKE 'test:*') AND _value > 10) OR ((_key LIKE 'test:*') AND _value < 5)",
				"(_key LIKE 'test:*') AND ((_value > 10) OR (_value < 5))",
				"(_value > 10 AND _key LIKE 'test:*') OR (_value < 5 AND _key LIKE 'test:*')",
				"(_value < 5 AND _key LIKE 'test:*') OR (_value > 10 AND _key LIKE 'test:*')",
			},
			fingerprint: "f_6936111135456499050",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, query := range tt.similarEq {
				statement, err := sqlparser.Parse("SELECT * WHERE " + query)
				if err != nil {
					t.Fail()
				}
				selectStmt, ok := statement.(*sqlparser.Select)
				if !ok {
					t.Fail()
				}
				assert.DeepEqual(t, tt.fingerprint, sql.GenerateFingerprint(selectStmt.Where.Expr))
			}
		})
	}
}
