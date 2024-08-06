package tests

import (
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

type TestCase struct {
	name     string
	commands []string
	expected []interface{}
}

func TestSet(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{
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
}

func TestSetWithOptions(t *testing.T) {
	conn := getLocalConnection()
	expiryTime := strconv.FormatInt(time.Now().Add(1*time.Minute).UnixMilli(), 10)
	defer conn.Close()

	testCases := []TestCase{
		{
			name:     "Set with EX option",
			commands: []string{"SET k v EX 2", "GET k", "SLEEP 3", "GET k"},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name:     "Set with PX option",
			commands: []string{"SET k v PX 2000", "GET k", "SLEEP 3", "GET k"},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name:     "Set with EX and PX option",
			commands: []string{"SET k v EX 2 PX 2000"},
			expected: []interface{}{"ERR syntax error"},
		},
		{
			name:     "XX on non-existing key",
			commands: []string{"DEL k", "SET k v XX", "GET k"},
			expected: []interface{}{int64(0), "(nil)", "(nil)"},
		},
		{
			name:     "NX on non-existing key",
			commands: []string{"DEL k", "SET k v NX", "GET k"},
			expected: []interface{}{int64(0), "OK", "v"},
		},
		{
			name:     "NX on existing key",
			commands: []string{"DEL k", "SET k v NX", "GET k", "SET k v NX"},
			expected: []interface{}{int64(0), "OK", "v", "(nil)"},
		},
		{
			name:     "PXAT option",
			commands: []string{"SET k v PXAT " + expiryTime, "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "PXAT option with delete",
			commands: []string{"SET k1 v1 PXAT " + expiryTime, "GET k1", "SLEEP 2", "DEL k1"},
			expected: []interface{}{"OK", "v1", "OK", int64(1)},
		},
		{
			name:     "PXAT option with invalid unix time ms",
			commands: []string{"SET k2 v2 PXAT 123123", "GET k2"},
			expected: []interface{}{"OK", "(nil)"},
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
			deleteTestKeys([]string{"k", "k1", "k2"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithExat(t *testing.T) {
	conn := getLocalConnection()
	Etime := strconv.FormatInt(time.Now().Unix()+10, 10)
	for _, tcase := range []TestCase{
		{
			commands: []string{"SET k v EXAT " + Etime, "TTL k"},
			expected: []interface{}{"OK", int64(10)},
		},
	} {
		for i := 0; i < len(tcase.commands); i++ {
			cmd := tcase.commands[i]
			out := tcase.expected[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}
}

func TestSetWithKeepTTLFlag(t *testing.T) {
	conn := getLocalConnection()
	for _, tcase := range []TestCase{
		{
			commands: []string{"SET k v EX 2", "SET k vv KEEPTTL", "GET k"},
			expected: []interface{}{"OK", "OK", "vv"},
		},
	} {
		for i := 0; i < len(tcase.commands); i++ {
			cmd := tcase.commands[i]
			out := tcase.expected[i]
			assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}

	time.Sleep(2 * time.Second)

	cmd := "GET k"
	out := "(nil)"

	assert.Equal(t, out, fireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
}
