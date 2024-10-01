package websocket

import (
	"fmt"
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
	exec := NewWebsocketCommandExecutor()

	testCases := []TestCase{
		{
			name:     "Set and Get Simple Value",
			commands: []string{"SET k v", "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "Set and Get Integer Value",
			commands: []string{"SET k 123456789", "GET k"},
			expected: []interface{}{"OK", float64(123456789)},
		},
		{
			name:     "Overwrite Existing Key",
			commands: []string{"SET k v1", "SET k 5", "GET k"},
			expected: []interface{}{"OK", "OK", float64(5)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()
			exec.FireCommand(conn, "del k")

			for i, cmd := range tc.commands {
				result := exec.FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithOptions(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	expiryTime := strconv.FormatInt(time.Now().Add(1*time.Minute).UnixMilli(), 10)

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
			expected: []interface{}{float64(0), "(nil)", "(nil)"},
		},
		{
			name:     "NX on non-existing key",
			commands: []string{"DEL k", "SET k v NX", "GET k"},
			expected: []interface{}{float64(0), "OK", "v"},
		},
		{
			name:     "NX on existing key",
			commands: []string{"DEL k", "SET k v NX", "GET k", "SET k v NX"},
			expected: []interface{}{float64(0), "OK", "v", "(nil)"},
		},
		{
			name:     "PXAT option",
			commands: []string{"SET k v PXAT " + expiryTime, "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "PXAT option with delete",
			commands: []string{"SET k1 v1 PXAT " + expiryTime, "GET k1", "SLEEP 2", "DEL k1"},
			expected: []interface{}{"OK", "v1", "OK", float64(1)},
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
			conn := exec.ConnectToServer()
			exec.FireCommand(conn, "del k")
			exec.FireCommand(conn, "del k1")
			exec.FireCommand(conn, "del k2")
			for i, cmd := range tc.commands {
				result := exec.FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithExat(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	Etime := strconv.FormatInt(time.Now().Unix()+5, 10)
	BadTime := "123123"

	t.Run("SET with EXAT",
		func(t *testing.T) {
			conn := exec.ConnectToServer()
			exec.FireCommand(conn, "DEL k")
			assert.Equal(t, "OK", exec.FireCommand(conn, fmt.Sprintf("SET k v EXAT %v", Etime)), "Value mismatch for cmd SET k v EXAT "+Etime)
			assert.Equal(t, "v", exec.FireCommand(conn, "GET k"), "Value mismatch for cmd GET k")
			assert.Assert(t, exec.FireCommand(conn, "TTL k").(float64) <= 5, "Value mismatch for cmd TTL k")
			time.Sleep(3 * time.Second)
			assert.Assert(t, exec.FireCommand(conn, "TTL k").(float64) <= 3, "Value mismatch for cmd TTL k")
			time.Sleep(3 * time.Second)
			assert.Equal(t, "(nil)", exec.FireCommand(conn, "GET k"), "Value mismatch for cmd GET k")
			assert.Equal(t, float64(-2), exec.FireCommand(conn, "TTL k"), "Value mismatch for cmd TTL k")
		})

	t.Run("SET with invalid EXAT expires key immediately",
		func(t *testing.T) {
			conn := exec.ConnectToServer()
			exec.FireCommand(conn, "DEL k")
			assert.Equal(t, "OK", exec.FireCommand(conn, "SET k v EXAT "+BadTime), "Value mismatch for cmd SET k v EXAT "+BadTime)
			assert.Equal(t, "(nil)", exec.FireCommand(conn, "GET k"), "Value mismatch for cmd GET k")
			assert.Equal(t, float64(-2), exec.FireCommand(conn, "TTL k"), "Value mismatch for cmd TTL k")
		})

	t.Run("SET with EXAT and PXAT returns syntax error",
		func(t *testing.T) {
			conn := exec.ConnectToServer()
			exec.FireCommand(conn, "DEL k")
			assert.Equal(t, "ERR syntax error", exec.FireCommand(conn, "SET k v PXAT "+Etime+" EXAT "+Etime), "Value mismatch for cmd SET k v PXAT "+Etime+" EXAT "+Etime)
			assert.Equal(t, "(nil)", exec.FireCommand(conn, "GET k"), "Value mismatch for cmd GET k")
		})
}

func TestWithKeepTTLFlag(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	for _, tcase := range []TestCase{
		{
			commands: []string{"SET k v EX 2", "SET k vv KEEPTTL", "GET k", "SET kk vv", "SET kk vvv KEEPTTL", "GET kk"},
			expected: []interface{}{"OK", "OK", "vv", "OK", "OK", "vvv"},
		},
	} {
		for i := 0; i < len(tcase.commands); i++ {
			cmd := tcase.commands[i]
			out := tcase.expected[i]
			assert.Equal(t, out, exec.FireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}

	time.Sleep(2 * time.Second)

	cmd := "GET k"
	out := "(nil)"

	assert.Equal(t, out, exec.FireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
}
