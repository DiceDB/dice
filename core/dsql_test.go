package core

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
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
			sql:  "SELECT $key, $value FROM `match:100:*` ORDER BY $value DESC LIMIT 10",
			want: DSQLQuery{
				Selection:    QuerySelection{KeySelection: true, ValueSelection: true},
				RegexMatcher: "match:100:*",
				OrderBy:      QueryOrder{OrderBy: "$value", Order: "desc"},
				Limit:        10,
			},
			wantErr: false,
		},
		{
			name: "valid select key only with order and limit",
			sql:  "SELECT $key FROM `match:100:*` ORDER BY $key LIMIT 5",
			want: DSQLQuery{
				Selection:    QuerySelection{KeySelection: true, ValueSelection: false},
				RegexMatcher: "match:100:*",
				OrderBy:      QueryOrder{OrderBy: "$key", Order: "asc"},
				Limit:        5,
			},
			wantErr: false,
		},
		{
			name:    "invalid multiple fields",
			sql:     "SELECT field1, field2",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "only $key and $value are supported in SELECT expressions",
		},
		{
			name:    "invalid non-select statement",
			sql:     "INSERT INTO table_name (field_name) values ('value')",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "unsupported DSQL statement: *sqlparser.Insert",
		},
		{
			name:    "empty invalid statement",
			sql:     "",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "error parsing SQL statement: syntax error at position 1",
		},
		{
			name:    "unsupported having clause",
			sql:     "SELECT $key FROM `match:100:*` HAVING $key > 1",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "HAVING and GROUP BY clauses are not supported",
		},
		{
			name:    "unsupported group by clause",
			sql:     "SELECT $key FROM `match:100:*` GROUP BY $key",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "HAVING and GROUP BY clauses are not supported",
		},
		{
			name:    "invalid limit value",
			sql:     "SELECT $key FROM `match:100:*` LIMIT abc",
			want:    DSQLQuery{},
			wantErr: true,
			error:   "invalid LIMIT value",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := ParseQuery(testCase.sql)
			if testCase.wantErr {
				assert.Error(t, err, testCase.error, "ParseQuery() should have returned an error")
			} else {
				assert.NilError(t, err, "ParseQuery() should not have returned an error")
				assert.Check(t, cmp.DeepEqual(testCase.want, got), "ParseQuery() did not return expected output")
			}
		})
	}
}
