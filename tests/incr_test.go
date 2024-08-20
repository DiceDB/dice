package tests

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/dicedb/dice/internal/constants"
	"gotest.tools/v3/assert"
)

func TestINCR(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []struct {
			op          string
			key         string
			val         int64
			expectedErr string
		}
	}{
		{
			name: "Increment multiple keys",
			commands: []struct {
				op          string
				key         string
				val         int64
				expectedErr string
			}{
				{"s", "key1", 0, constants.EmptyStr},
				{"i", "key1", 1, constants.EmptyStr},
				{"i", "key1", 2, constants.EmptyStr},
				{"i", "key2", 1, constants.EmptyStr},
				{"g", "key1", 2, constants.EmptyStr},
				{"g", "key2", 1, constants.EmptyStr},
				{"s", "key3", math.MaxInt64 - 1, constants.EmptyStr},
				{"i", "key3", math.MaxInt64, constants.EmptyStr},
				{"i", "key3", math.MaxInt64, "ERR value is out of range"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"key1", "key2"}, store)
			fireCommand(conn, "DEL key1")
			fireCommand(conn, "DEL key2")

			for _, cmd := range tc.commands {
				switch cmd.op {
				case "s":
					fireCommand(conn, fmt.Sprintf("SET %s %d", cmd.key, cmd.val))
				case "i":
					result := fireCommand(conn, fmt.Sprintf("INCR %s", cmd.key))
					switch v := result.(type) {
					case string:
						assert.Equal(t, cmd.expectedErr, v)
					case int64:
						assert.Equal(t, cmd.val, v)
					}
				case "g":
					result := fireCommand(conn, fmt.Sprintf("GET %s", cmd.key))
					assert.Equal(t, strconv.FormatInt(cmd.val, 10), result)
				}
			}
		})
	}
}
