package tests

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
)

func TestDECR(t *testing.T) {
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
			name: "Decrement multiple keys",
			commands: []struct {
				op          string
				key         string
				val         int64
				expectedErr string
			}{
				{"s", "key1", 3, ""},
				{"d", "key1", 2, ""},
				{"d", "key1", 1, ""},
				{"d", "key2", -1, ""},
				{"g", "key1", 1, ""},
				{"g", "key2", -1, ""},
				{"s", "key3", math.MinInt64 + 1, ""},
				{"d", "key3", math.MinInt64, ""},
				{"d", "key3", math.MinInt64, "ERR value is out of range"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.commands {
				switch cmd.op {
				case "s":
					fireCommand(conn, fmt.Sprintf("SET %s %d", cmd.key, cmd.val))
				case "d":
					result := fireCommand(conn, fmt.Sprintf("DECR %s", cmd.key))
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
