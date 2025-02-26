package ironhawk

import (
	"strconv"
	"testing"
	"time"
)

func TestEXPIRETIME(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	futureUnixTimestamp := time.Now().Unix() + 1

	testCases := []TestCase{
		{
			name: "EXPIRETIME command",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key " + strconv.FormatInt(futureUnixTimestamp, 10),
				"EXPIRETIME test_key",
			},
			expected: []interface{}{"OK", 1, futureUnixTimestamp},
		},
		{
			name: "EXPIRETIME non-existent key",
			commands: []string{
				"EXPIRETIME non_existent_key",
			},
			expected: []interface{}{int64(-2)},
		},
		{
			name: "EXPIRETIME with past time",
			commands: []string{
				"SET test_key test_value",
				"EXPIREAT test_key 1724167183",
				"EXPIRETIME test_key",
			},
			expected: []interface{}{"OK", 1, int64(-2)},
		},
		{
			name: "EXPIRETIME with invalid syntax",
			commands: []string{
				"SET test_key test_value",
				"EXPIRETIME",
				"EXPIRETIME key1 key2",
			},
			expected: []interface{}{"OK", "ERR wrong number of arguments for 'expiretime' command", "ERR wrong number of arguments for 'expiretime' command"},
		},
	}
	runTestcases(t, client, testCases)
}
