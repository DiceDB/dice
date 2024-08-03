package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

type commandGetKeysTestCase struct {
	inCmd    string
	expected interface{}
}

func TestCommandGetKeys(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tcase := range []commandGetKeysTestCase{
		{"set 1 2 3 4", []interface{}{"1"}},
		{"get key", []interface{}{"key"}},
		{"ttl key", []interface{}{"key"}},
		{"del 1 2 3 4 5 6", []interface{}{"1", "2", "3", "4", "5", "6"}},
		{"expire key time extra", []interface{}{"key"}},
		{"QINTINS k 1", []interface{}{"k"}},
		{"BFINIT bloom some parameters", []interface{}{"bloom"}},

		{"ping", "ERR the command has no key arguments"},
		{"get", "ERR invalid number of arguments specified for command"},
		{"abort", "ERR the command has no key arguments"},
		{"NotValidCommand", "ERR invalid command specified"},
	} {
		cmd := tcase.inCmd
		out := tcase.expected
		result := fireCommand(conn, "COMMAND GETKEYS "+cmd)
		assert.DeepEqual(t, out, result)
	}
}
