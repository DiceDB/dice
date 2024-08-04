package tests

import (
	"fmt"
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
)

func TestINCR(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []struct {
			op  string
			key string
			val int64
		}
	}{
		{
			name: "Increment multiple keys",
			commands: []struct {
				op  string
				key string
				val int64
			}{
				{"s", "key1", 0},
				{"i", "key1", 1},
				{"i", "key1", 2},
				{"i", "key2", 1},
				{"g", "key1", 2},
				{"g", "key2", 1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.commands {
				switch cmd.op {
				case "s":
					fireCommand(conn, fmt.Sprintf("SET %s %d", cmd.key, cmd.val))
				case "i":
					fireCommand(conn, fmt.Sprintf("INCR %s", cmd.key))
				case "g":
					result := fireCommand(conn, fmt.Sprintf("GET %s", cmd.key))
					assert.Equal(t, strconv.FormatInt(cmd.val, 10), result)
				}
			}
		})
	}
}
