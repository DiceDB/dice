package tests

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestSet(t *testing.T) {
	conn := getLocalConnection()

	t.Run("Basic SET and GET", func(t *testing.T) {
		testCases := []struct {
			name     string
			commands []string
			expected []interface{}
		}{
			{
				name:     "Set and Get Simple Value",
				commands: []string{"SET k v", "GET k"},
				expected: []interface{}{"OK", "v"},
			},
			{
				name:     "Overwrite Existing Key",
				commands: []string{"SET k v1", "SET k v2", "GET k"},
				expected: []interface{}{"OK", "OK", "v2"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				deleteTestKeys([]string{"k"})
				for i, cmd := range tc.commands {
					result := fireCommand(conn, cmd)
					assert.DeepEqual(t, tc.expected[i], result)
				}
			})
		}
	})
}

func TestSetWithOptions(t *testing.T) {
	conn := getLocalConnection()

	t.Run("SET with XX option", func(t *testing.T) {
		testCases := []struct {
			name     string
			commands []string
			expected []interface{}
		}{
			{
				name:     "XX on non-existing key",
				commands: []string{"DEL k", "SET k v XX", "GET k"},
				expected: []interface{}{int64(0), "(nil)", "(nil)"},
			},
			{
				name:     "XX on existing key",
				commands: []string{"SET k v1", "SET k v2 XX", "GET k"},
				expected: []interface{}{"OK", "OK", "v2"},
			},
			{
				name:     "Multiple XX operations",
				commands: []string{"SET k v1", "SET k v2 XX", "SET k v3 XX", "GET k"},
				expected: []interface{}{"OK", "OK", "OK", "v3"},
			},
			{
				name:     "EX option",
				commands: []string{"SET k v EX 1", "GET k", "SLEEP 2", "GET k"},
				expected: []interface{}{"OK", "v", "OK", "(nil)"},
			},
			{
				name:     "XX option",
				commands: []string{"SET k v XX EX 1", "GET k", "SLEEP 2", "GET k", "SET k v XX EX 1", "GET k"},
				expected: []interface{}{"(nil)", "(nil)", "OK", "(nil)", "(nil)", "(nil)"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				deleteTestKeys([]string{"k"})
				for i, cmd := range tc.commands {
					result := fireCommand(conn, cmd)
					assert.Equal(t, tc.expected[i], result)
				}
			})
		}
	})
}

func TestSetWithKeepTTLFlag(t *testing.T) {
	conn := getLocalConnection()
	for _, tcase := range []DTestCase{
		{
			InCmds: []string{"SET k v EX 5", "SET k vv KEEPTTL", "GET k"},
			Out:    []interface{}{"OK", "OK", "vv"},
		},
	} {
		for i := 0; i < len(tcase.InCmds); i++ {
			cmd := tcase.InCmds[i]
			out := tcase.Out[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}

	time.Sleep(5 * time.Second)
	// cmd := "GET k"
	// out := "nil"
	// assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
}
