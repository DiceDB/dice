package sql_test

import (
	"testing"

	"github.com/dicedb/dice/internal/sql"
	"github.com/xwb1989/sqlparser"
	"gotest.tools/v3/assert"
)

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
