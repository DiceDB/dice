package commands

import (
	"encoding/base64"
	"fmt"

	"testing"

	"github.com/dicedb/dice/testutils"
	"gotest.tools/v3/assert"
)


func TestDump(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "DUMP string value",
			commands: []string{
				"SET mykey hello",
				"DUMP mykey",
				"DEL mykey",
				"RESTORE mykey abc",
				"GET mykey",
			},
			expected: []interface{}{
				"OK",
				func(result interface{}) bool {
					dumped, ok := result.(string)
					if !ok {
						return false
					}
					decoded, err := base64.StdEncoding.DecodeString(dumped)
					if err != nil {
						return false
					}
					return len(decoded) > 11 &&
						decoded[0] == 0x09 && 
						decoded[1] == 0x00 && 
						string(decoded[6:11]) == "hello" &&
						decoded[11] == 0xFF
				},
				int64(1),
				"OK",
				"OK",
			},
		},
		{
			name: "DUMP non-existent key",
			commands: []string{
				"DUMP nonexistentkey",
			},
			expected: []interface{}{
				"ERR could not perform this operation on a key that doesn't exist",
			},
		},
		{
			name: "DUMP integer value",
			commands: []string{
				"SET intkey 42",
				"DUMP intkey",
			},
			expected: []interface{}{
				"OK",
				func(result interface{}) bool {
					dumped, ok := result.(string)
					if !ok {
						return false
					}
					decoded, err := base64.StdEncoding.DecodeString(dumped)
					if err != nil {
						return false
					}
					return len(decoded) > 2 &&
						decoded[0] == 0x09 &&
						decoded[1] == 0xC0
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "FLUSHALL")
			for i, cmd := range tc.commands {
				var result interface{}
				result = FireCommand(conn, cmd)
				fmt.Println(cmd)
				expected := tc.expected[i]
				fmt.Println(tc.expected)
				fmt.Println(result)
				switch exp := expected.(type) {
				case string:
					assert.DeepEqual(t, exp, result)
				case []interface{}:
					assert.Assert(t, testutils.UnorderedEqual(exp, result))
				case func(interface{}) bool:
					assert.Assert(t, exp(result),cmd)
				default:
					assert.DeepEqual(t, expected, result)
				}
			}
		})
	}
}