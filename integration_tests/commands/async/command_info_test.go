package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var getInfoTestCases = []struct {
	name     string
	inCmd    string
	expected interface{}
}{
	{"Set command", "SET", []interface{}{[]interface{}{"set", int64(-3), int64(1), int64(0), int64(0), []any{}}}},
	{"Get command", "GET", []interface{}{[]interface{}{"get", int64(2), int64(1), int64(0), int64(0), []any{}}}},
	{"Ping command", "PING", []interface{}{[]interface{}{"ping", int64(-1), int64(0), int64(0), int64(0), []any{}}}},
	{"Invalid command", "INVALID_CMD", []interface{}{string("(nil)")}},
	{"Combination of valid and Invalid command", "SET INVALID_CMD", []interface{}{
		[]interface{}{"set", int64(-3), int64(1), int64(0), int64(0), []any{}},
		string("(nil)"),
	}},
	{"Combination of multiple valid commands", "SET GET", []interface{}{
		[]interface{}{"set", int64(-3), int64(1), int64(0), int64(0), []any{}},
		[]interface{}{"get", int64(2), int64(1), int64(0), int64(0), []any{}},
	}},
}

func TestCommandInfo(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range getInfoTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FireCommand(conn, "COMMAND INFO "+tc.inCmd)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func BenchmarkCommandInfo(b *testing.B) {
	conn := getLocalConnection()
	defer conn.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range getKeysTestCases {
			FireCommand(conn, "COMMAND INFO "+tc.inCmd)
		}
	}
}
