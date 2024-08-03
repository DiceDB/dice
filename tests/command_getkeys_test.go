package tests

import (
	"testing"
)

type commandGetKeysTestCase struct {
	inCmd    string
	expected interface{}
}

var getKeysTestCases = []commandGetKeysTestCase{
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
}

// func TestCommandGetKeys(t *testing.T) {
// 	conn := getLocalConnection()
// 	defer conn.Close()

// 	for _, tcase := range getKeysTestCases {
// 		fmt.Println("done")
// 		result := fireCommand(conn, "COMMAND GETKEYS "+tcase.inCmd)
// 		assert.DeepEqual(t, tcase.expected, result)
// 	}
// }

func BenchmarkGetKeysMatch(b *testing.B) {
	conn := getLocalConnection()
	for _, tcase := range getKeysTestCases {
		for i := 0; i < b.N; i++ {
			fireCommand(conn, "COMMAND GETKEYS "+tcase.inCmd)
		}
	}
}
