package sql

import (
	"strings"
	"testing"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/stretchr/testify/assert"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    DSQLQuery
		wantErr bool
		error   string
	}{
		{
			name: "valid select key and value with order and limit",
			sql:  "SELECT $key, $value WHERE $key like `match:100:*` ORDER BY $value DESC LIMIT 10",
			want: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: true},
				Where: &ComparisonNode{
					Operator: "like",
					Left:     NewKeyType(),
					Right:    NewStringType("match:100:*"),
				},
				OrderBy: QueryOrder{OrderBy: Value{Value: "_value", Type: FieldVal}, Order: "desc"},
				Limit:   10,
			},
			wantErr: false,
		},
		{
			name: "valid select with where clause",
			sql:  "SELECT $key, $value WHERE $key like `match:100:*` AND $value = 'test' ORDER BY $key LIMIT 5",
			want: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: true},
				Where: &AndNode{
					Left: &ComparisonNode{
						Operator: "like",
						Left:     NewKeyType(),
						Right:    NewStringType("match:100:*"),
					},
					Right: &ComparisonNode{
						Operator: "=",
						Left:     NewValueType(),
						Right:    NewStringType("test"),
					},
				},
				OrderBy: QueryOrder{OrderBy: Value{Value: "_key", Type: FieldKey}, Order: Asc},
				Limit:   5,
			},
			wantErr: false,
		},
		{
			name: "complex where clause",
			sql:  "SELECT $key WHERE $key like `user:*` AND $value > 25 AND $key LIKE 'user:1%'",
			want: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: false},
				Where: &AndNode{
					Left: &AndNode{
						Left: &ComparisonNode{
							Operator: "like",
							Left:     NewKeyType(),
							Right:    NewStringType("user:*"),
						},
						Right: &ComparisonNode{
							Operator: ">",
							Left:     NewValueType(),
							Right:    NewIntType("25"),
						},
					},
					Right: &ComparisonNode{
						Operator: "like",
						Left:     NewKeyType(),
						Right:    NewStringType("user:1%"),
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid order by expression",
			sql:     "SELECT $key WHERE $key like `match:100:*` ORDER BY invalid_key LIMIT 5",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:51 - mismatched input 'invalid_key'",
		},
		{
			name:    "invalid multiple fields",
			sql:     "SELECT field1, field2 WHERE $key like `test`",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:7 - mismatched input 'field1'",
		},
		{
			name:    "invalid non-select statement",
			sql:     "INSERT INTO table_name (field_name) values ('value')",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:0 - mismatched input 'INSERT' expecting 'SELECT'",
		},
		{
			name:    "empty invalid statement",
			sql:     utils.EmptyStr,
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:0 - mismatched input '<EOF>' expecting 'SELECT'",
		},
		{
			name:    "unsupported having clause",
			sql:     "SELECT $key WHERE $key like `match:100:*` HAVING $key > 1",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:42 - mismatched input 'HAVING' expecting <EOF>",
		},
		{
			name:    "unsupported group by clause",
			sql:     "SELECT $key WHERE $key like `match:100:*` GROUP BY $key",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:42 - mismatched input 'GROUP' expecting <EOF>",
		},
		{
			name:    "invalid limit value",
			sql:     "SELECT $key WHERE $key like `match:100:*` LIMIT abc",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:48 - mismatched input 'abc' expecting NUMBER",
		},
		{
			name: "select only value",
			sql:  "SELECT $value WHERE $key like `test:*`",
			want: DSQLQuery{
				Selection: QuerySelection{KeySelection: false, ValueSelection: true},
				Where: &ComparisonNode{
					Operator: "like",
					Left:     NewKeyType(),
					Right:    NewStringType("test:*"),
				},
			},
			wantErr: false,
		},
		{
			name: "order by key ascending",
			sql:  "SELECT $key, $value WHERE $key like `test:*` ORDER BY $key ASC",
			want: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: true},
				Where: &ComparisonNode{
					Operator: "like",
					Left:     NewKeyType(),
					Right:    NewStringType("test:*"),
				},
				OrderBy: QueryOrder{OrderBy: Value{Value: "_key", Type: FieldKey}, Order: "asc"},
			},
			wantErr: false,
		},
		{
			name:    "invalid table name",
			sql:     "SELECT $key FROM 123",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:12 - mismatched input 'FROM' expecting <EOF>",
		},
		{
			name:    "Banned FROM clause",
			sql:     "SELECT $key FROM tablename",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "syntax error at line 1:12 - mismatched input 'FROM' expecting <EOF>",
		},
		{
			name: "where clause with NULL comparison",
			sql:  "SELECT $key, $value WHERE $key like `test:*` AND $value IS NULL",
			want: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: true},
				Where: &AndNode{
					Left: &ComparisonNode{
						Operator: "like",
						Left:     NewKeyType(),
						Right:    NewStringType("test:*"),
					},
					Right: &ComparisonNode{
						Operator: "is",
						Left:     NewValueType(),
						Right:    NewStringType("NULL"), // FIX: Null type
					},
				},
			},
			wantErr: false,
		},
		{
			name: "where clause with multiple conditions",
			sql:  "SELECT $key WHERE ($key LIKE `test:*`) AND ($value > 10 OR $value < 5)",
			want: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: false},
				Where: &AndNode{
					Left: &ComparisonNode{
						Operator: "like",
						Left:     NewKeyType(),
						Right:    NewStringType("test:*"),
					},
					Right: &OrNode{
						Left: &ComparisonNode{
							Operator: ">",
							Left:     NewValueType(),
							Right:    NewIntType("10"),
						},
						Right: &ComparisonNode{
							Operator: "<",
							Left:     NewValueType(),
							Right:    NewIntType("5"),
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuery(tt.sql)
			if tt.wantErr {
				assert.ErrorContains(t, err, tt.error)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want.Selection, got.Selection)
				assert.Equal(t, tt.want.OrderBy, got.OrderBy)
				assert.Equal(t, tt.want.Limit, got.Limit)

				//if tt.want.Where == nil {
				//	assert.Assert(t, got.Where == nil)
				//} else {
				assert.True(t, got.Where != nil)
				assert.Equal(t, tt.want.Where, got.Where)
				//}
			}
		})
	}
}

func TestParseSelectExpressions(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    QuerySelection
		wantErr bool
	}{
		{
			name: "select key and value",
			sql:  "SELECT $key, $value WHERE $key like `test`",
			want: QuerySelection{KeySelection: true, ValueSelection: true},
		},
		{
			name: "select only key",
			sql:  "SELECT $key WHERE $key like `test`",
			want: QuerySelection{KeySelection: true, ValueSelection: false},
		},
		{
			name: "select only value",
			sql:  "SELECT $value WHERE $key like `test`",
			want: QuerySelection{KeySelection: false, ValueSelection: true},
		},
		{
			name:    "select invalid field",
			sql:     "SELECT invalid WHERE $key like `test`",
			wantErr: true,
		},
		{
			name:    "select too many fields",
			sql:     "SELECT $key, $value, extra WHERE $key like `test`",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuery(tt.sql)

			if tt.wantErr {
				assert.True(t, err != nil)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got.Selection)
			}
		})
	}
}

func TestParseOrderBy(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    QueryOrder
		wantErr bool
	}{
		{
			name: "order by key asc",
			sql:  "SELECT $key WHERE $key like `test` ORDER BY $key ASC",
			want: QueryOrder{OrderBy: Value{Value: "_key", Type: FieldKey}, Order: Asc},
		},
		{
			name: "order by key desc",
			sql:  "SELECT $key WHERE $key like `test` ORDER BY $key DESC",
			want: QueryOrder{OrderBy: Value{Value: "_key", Type: FieldKey}, Order: "desc"},
		},
		{
			name: "order by value asc",
			sql:  "SELECT $value WHERE $key like `test` ORDER BY $value ASC",
			want: QueryOrder{OrderBy: Value{Value: "_value", Type: FieldVal}, Order: "asc"},
		},
		{
			name: "order by value desc",
			sql:  "SELECT $value WHERE $key like `test` ORDER BY $value DESC",
			want: QueryOrder{OrderBy: Value{Value: "_value", Type: FieldVal}, Order: "desc"},
		},
		{
			name: "order by json path asc",
			sql:  "SELECT $value WHERE $key like `test` ORDER BY $value.name ASC",
			want: QueryOrder{OrderBy: Value{Value: "_value.name", Type: FieldJSON}, Order: "asc"},
		},
		{
			name: "order by nested json path desc",
			sql:  "SELECT $value WHERE $key like `test` ORDER BY $value.address.city DESC",
			want: QueryOrder{OrderBy: Value{Value: "_value.address.city", Type: FieldJSON}, Order: "desc"},
		},
		{
			name: "order by json path with array index",
			sql:  "SELECT $value WHERE $key like `test` ORDER BY `$value.items[0].price`",
			want: QueryOrder{OrderBy: Value{Value: "_value.items[0].price", Type: FieldJSON}, Order: "asc"},
		},
		{
			name: "order by complex json path",
			sql:  "SELECT $value WHERE $key like `test` ORDER BY `$value.users[*].contacts[0].email`",
			want: QueryOrder{OrderBy: Value{Value: "_value.users[*].contacts[0].email", Type: FieldJSON}, Order: "asc"},
		},
		{
			name: "no order by clause",
			sql:  "SELECT $key WHERE $key like `test`",
			want: QueryOrder{},
		},
		{
			name:    "invalid order by field",
			sql:     "SELECT $key WHERE $key like `test` ORDER BY invalid",
			wantErr: true,
		},
		{
			name: "no order by clause",
			sql:  "SELECT $key WHERE $key like `test`",
			want: QueryOrder{},
		},
		{
			name:    "multiple order by clauses",
			sql:     "SELECT $key WHERE $key like `test` ORDER BY $key ASC, $value DESC",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuery(tt.sql)

			if tt.wantErr {
				assert.True(t, err != nil)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got.OrderBy)
			}
		})
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    int
		wantErr bool
	}{
		{
			name: "valid limit",
			sql:  "SELECT $key WHERE $key like `test` LIMIT 10",
			want: 10,
		},
		{
			name: "no limit clause",
			sql:  "SELECT $key WHERE $key like `test`",
			want: 0,
		},
		{
			name:    "invalid limit value",
			sql:     "SELECT $key WHERE $key like `test` LIMIT abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQuery(tt.sql)

			if tt.wantErr {
				assert.True(t, err != nil)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got.Limit)
			}
		})
	}
}

func TestDSQLQueryString(t *testing.T) {
	tests := []struct {
		name     string
		query    DSQLQuery
		expected string
	}{
		{
			name: "Key Selection Only",
			query: DSQLQuery{
				Selection: QuerySelection{KeySelection: true},
			},
			expected: "SELECT $key",
		},
		{
			name: "Value Selection Only",
			query: DSQLQuery{
				Selection: QuerySelection{ValueSelection: true},
			},
			expected: "SELECT $value",
		},
		{
			name: "Both Key and Value Selection",
			query: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: true},
			},
			expected: "SELECT $key, $value",
		},
		{
			query: DSQLQuery{
				Selection: QuerySelection{ValueSelection: true},
			},
			expected: "SELECT $value",
		},
		{
			name: "With Where Clause",
			query: DSQLQuery{
				Selection: QuerySelection{KeySelection: true},
				Where: &ComparisonNode{
					Operator: ">",
					Left:     NewValueType(),
					Right:    NewIntType("10"),
				},
			},
			expected: "SELECT $key WHERE $value > 10",
		},
		{
			name: "With OrderBy",
			query: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: true},
				OrderBy:   QueryOrder{OrderBy: Value{Value: "_key", Type: FieldKey}, Order: "DESC"},
			},
			expected: "SELECT $key, $value ORDER BY $key DESC",
		},
		{
			name: "With Limit",
			query: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: true},
				Limit:     5,
			},
			expected: "SELECT $key, $value LIMIT 5",
		},
		{
			name: "Full Query",
			query: DSQLQuery{
				Selection: QuerySelection{KeySelection: true, ValueSelection: true},
				Where: &AndNode{
					Left: &ComparisonNode{
						Operator: "like",
						Left:     NewKeyType(),
						Right:    NewJSONType("match:100:*"),
					},
					Right: &ComparisonNode{
						Operator: "=",
						Left:     NewValueType(),
						Right:    NewStringType("test"),
					},
				},
				OrderBy: QueryOrder{OrderBy: Value{Value: "_key", Type: FieldKey}, Order: "DESC"},
				Limit:   5,
			},
			expected: "SELECT $key, $value WHERE $key like 'match:100:*' and $value = test ORDER BY $key DESC LIMIT 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.query.String()
			if !strings.EqualFold(result, tt.expected) {
				t.Errorf("Expected %q, but got %q", tt.expected, result)
			}
		})
	}
}
