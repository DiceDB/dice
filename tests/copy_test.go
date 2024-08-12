package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestCopy(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	simpleJSON := `{"name":"John","age":30}`

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "COPY when source key doesn't exist",
			commands: []string{"COPY k3 k2"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "COPY with no REPLACE",
			commands: []string{"SET k1 v1", "COPY k1 k2", "GET k1", "GET k2"},
			expected: []interface{}{"OK", int64(1), "v1", "v1"},
		},
		{
			name:     "COPY with REPLACE",
            commands: []string{"SET k4 v1", "SET k5 v2", "GET k5", "COPY k4 k5 REPLACE", "GET k5"},
			expected: []interface{}{"OK", "OK", "v2", int64(1), "v1"},
		},
        {
            name: "COPY with JSON integer",
            commands: []string{"JSON.SET k6 $ 2", "COPY k6 k7", "JSON.GET k7"},
            expected: []interface{}{"OK", int64(1), "2"},
        },
        {
            name: "COPY with JSON boolean",
            commands: []string{"JSON.SET k8 $ true", "COPY k8 k9", "JSON.GET k8"},
            expected: []interface{}{"OK", int64(1), "true"},
        },
        {
            name: "COPY with JSON array",
            commands: []string{`JSON.SET k10 $ [1,2,3]`, "COPY k10 k11", "JSON.GET k11"},
            expected: []interface{}{"OK", int64(1), `[1,2,3]`},
        },
        {
            name: "COPY with JSON simple JSON with REPLACE",
            commands: []string{`JSON.SET k12 $ ` + simpleJSON, "COPY k12 k11 REPLACE", "JSON.GET k11"},
            expected: []interface{}{"OK", int64(1), simpleJSON},
        },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k1", "k2", "k3", "k4", "k5", "k6", "k7", "k8", "k9", "k10", "k11"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
