package resp

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
			name:     "Set and Get Integer Value",
			commands: []string{"SET k 123456789", "GET k"},
			expected: []interface{}{"OK", int64(123456789)},
		},
		{
			name:     "Overwrite Existing Key",
			commands: []string{"SET k v1", "SET k 5", "GET k"},
			expected: []interface{}{"OK", "OK", int64(5)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"k"}, store)
			FireCommand(conn, "DEL k")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
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
			// deleteTestKeys([]string{"k", "k1", "k2"}, store)
			FireCommand(conn, "DEL k")
			FireCommand(conn, "DEL k1")
			FireCommand(conn, "DEL k2")
			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithExat(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	Etime := strconv.FormatInt(time.Now().Unix()+5, 10)
	BadTime := "123123"

	t.Run("SET with EXAT",
		func(t *testing.T) {
			// deleteTestKeys([]string{"k"}, store)
			FireCommand(conn, "DEL k")
			assert.Equal(t, "OK", FireCommand(conn, "SET k v EXAT "+Etime), "Value mismatch for cmd SET k v EXAT "+Etime)
			assert.Equal(t, "v", FireCommand(conn, "GET k"), "Value mismatch for cmd GET k")
			assert.True(t, FireCommand(conn, "TTL k").(int64) <= 5, "Value mismatch for cmd TTL k")
			time.Sleep(3 * time.Second)
			assert.True(t, FireCommand(conn, "TTL k").(int64) <= 3, "Value mismatch for cmd TTL k")
			time.Sleep(3 * time.Second)
			assert.Equal(t, "(nil)", FireCommand(conn, "GET k"), "Value mismatch for cmd GET k")
			assert.Equal(t, int64(-2), FireCommand(conn, "TTL k"), "Value mismatch for cmd TTL k")
		})

	t.Run("SET with invalid EXAT expires key immediately",
		func(t *testing.T) {
			// deleteTestKeys([]string{"k"}, store)
			FireCommand(conn, "DEL k")
			assert.Equal(t, "OK", FireCommand(conn, "SET k v EXAT "+BadTime), "Value mismatch for cmd SET k v EXAT "+BadTime)
			assert.Equal(t, "(nil)", FireCommand(conn, "GET k"), "Value mismatch for cmd GET k")
			assert.Equal(t, int64(-2), FireCommand(conn, "TTL k"), "Value mismatch for cmd TTL k")
		})

	t.Run("SET with EXAT and PXAT returns syntax error",
		func(t *testing.T) {
			// deleteTestKeys([]string{"k"}, store)
			FireCommand(conn, "DEL k")
			assert.Equal(t, "ERR syntax error", FireCommand(conn, "SET k v PXAT "+Etime+" EXAT "+Etime), "Value mismatch for cmd SET k v PXAT "+Etime+" EXAT "+Etime)
			assert.Equal(t, "(nil)", FireCommand(conn, "GET k"), "Value mismatch for cmd GET k")
		})
}

func TestWithKeepTTLFlag(t *testing.T) {
	conn := getLocalConnection()
	expiryTime := strconv.FormatInt(time.Now().Add(1*time.Minute).UnixMilli(), 10)
	defer conn.Close()

	for _, tcase := range []TestCase{
		{
			commands: []string{"SET k v EX 2", "SET k vv KEEPTTL", "GET k", "SET kk vv", "SET kk vvv KEEPTTL", "GET kk", "SET K V EX 2 KEEPTTL", "SET K1 vv PX 2000 KEEPTTL", "SET K2 vv EXAT " + expiryTime + " KEEPTTL"},
			expected: []interface{}{"OK", "OK", "vv", "OK", "OK", "vvv", "ERR syntax error", "ERR syntax error", "ERR syntax error"},
		},
	} {
		for i := 0; i < len(tcase.commands); i++ {
			cmd := tcase.commands[i]
			out := tcase.expected[i]
			assert.Equal(t, out, FireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
		}
	}

	time.Sleep(2 * time.Second)

	cmd := "GET k"
	out := "(nil)"

	assert.Equal(t, out, FireCommand(conn, cmd), "Value mismatch for cmd %s\n.", cmd)
}
