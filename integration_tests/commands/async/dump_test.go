package async

import (
	"encoding/base64"
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDumpRestore(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "DUMP and RESTORE string value",
			commands: []string{
				"SET mykey hello",
				"DUMP mykey",
				"DEL mykey",
				"RESTORE mykey 2 CQAAAAAFaGVsbG//AEeXk742Rcc=",
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
				"hello",
			},
		},
		{
			name: "DUMP and RESTORE integer value",
			commands: []string{
				"SET intkey 42",
				"DUMP intkey",
				"DEL intkey",
				"RESTORE intkey 2 CcAAAAAAAAAAKv9S/ymRDY3rXg==",
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
				int64(1),
				"OK",
			},
		},
		{
			name: "DUMP non-existent key",
			commands: []string{
				"DUMP nonexistentkey",
			},
			expected: []interface{}{
				"ERR nil",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "FLUSHALL")
			for i, cmd := range tc.commands {
				var result interface{}
				result = FireCommand(conn, cmd)
				expected := tc.expected[i]

				switch exp := expected.(type) {
				case string:
					assert.Equal(t, exp, result)
				case []interface{}:
					assert.True(t, testutils.UnorderedEqual(exp, result))
				case func(interface{}) bool:
					assert.True(t, exp(result), cmd)
				default:
					assert.Equal(t, expected, result)
				}
			}
		})
	}
}
