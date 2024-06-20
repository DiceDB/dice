package core

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    string
		wantErr bool
		error   string
	}{
		{
			name:    "valid select single field",
			sql:     "SELECT field_name",
			want:    "field_name",
			wantErr: false,
			error:   "",
		},
		{
			name:    "invalid multiple fields",
			sql:     "SELECT field1, field2",
			want:    "",
			wantErr: true,
			error:   "only single field selections are supported, found 2 fields",
		},
		{
			name:    "invalid non-select statement",
			sql:     "INSERT INTO table_name (field_name) VALUES ('value')",
			want:    "",
			wantErr: true,
			error:   "unsupported SQL statement: *sqlparser.Insert",
		},
		{
			name:    "invalid wildcard select",
			sql:     "SELECT * FROM table_name",
			want:    "",
			wantErr: true,
			error:   "only simple field selections are supported",
		},
		{
			name:    "empty invalid statement",
			sql:     "",
			want:    "",
			wantErr: true,
			error:   "error parsing SQL statement: syntax error at position 1",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := ParseQuery(testCase.sql)
			if testCase.wantErr {
				assert.Error(t, err, testCase.error, "ParseQuery() should have returned an error")
			} else {
				assert.NilError(t, err, "ParseQuery() should not have returned an error")
				assert.Equal(t, testCase.want, got, "ParseQuery() did not return expected output")
			}
		})
	}
}
