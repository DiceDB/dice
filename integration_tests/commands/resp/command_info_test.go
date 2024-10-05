package resp

import (
	"testing"

	"gotest.tools/v3/assert"
)

var getInfoTestCases = []struct {
	name     string
	inCmd    string
	expected interface{}
}{
	{"Set command", "SET", []interface{}{[]interface{}{"SET", int64(-3), int64(1), int64(0), int64(0)}}},
	{"Get command", "GET", []interface{}{[]interface{}{"GET", int64(2), int64(1), int64(0), int64(0)}}},
	{"Ping command", "PING", []interface{}{[]interface{}{"PING", int64(-1), int64(0), int64(0), int64(0)}}},
	{"Invalid command", "INVALID_CMD", []interface{}{string("(nil)")}},
	{"Combination of valid and Invalid command", "SET INVALID_CMD", []interface{}{
		[]interface{}{"SET", int64(-3), int64(1), int64(0), int64(0)},
		string("(nil)"),
	}},
	{"Combination of multiple valid commands", "SET GET", []interface{}{
		[]interface{}{"SET", int64(-3), int64(1), int64(0), int64(0)},
		[]interface{}{"GET", int64(2), int64(1), int64(0), int64(0)},
	}},
}

func TestCommandInfo(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range getInfoTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FireCommand(conn, "COMMAND INFO "+tc.inCmd)
			assert.DeepEqual(t, tc.expected, result)
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
