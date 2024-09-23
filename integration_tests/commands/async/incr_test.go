package async

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/server/utils"
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
			val         interface{}
			expectedErr string
		}
	}{
		{
			name: "Increment multiple keys",
			commands: []struct {
				op          string
				key         string
				val         interface{}
				expectedErr string
			}{
				{"s", "key1", int64(0), utils.EmptyStr},
				{"i", "key1", int64(1), utils.EmptyStr},
				{"i", "key1", int64(2), utils.EmptyStr},
				{"i", "key2", int64(1), utils.EmptyStr},
				{"g", "key1", int64(2), utils.EmptyStr},
				{"g", "key2", int64(1), utils.EmptyStr},
			},
		},
		{
			name: "Increment to and from max int64",
			commands: []struct {
				op          string
				key         string
				val         interface{}
				expectedErr string
			}{
				{"s", "max_int", int64(math.MaxInt64 - 1), utils.EmptyStr},
				{"i", "max_int", int64(math.MaxInt64), utils.EmptyStr},
				{"i", "max_int", nil, "ERR increment or decrement would overflow"},
				{"s", "max_int", int64(math.MaxInt64), utils.EmptyStr},
				{"i", "max_int", nil, "ERR increment or decrement would overflow"},
			},
		},
		{
			name: "Increment from min int64",
			commands: []struct {
				op          string
				key         string
				val         interface{}
				expectedErr string
			}{
				{"s", "min_int", int64(math.MinInt64), utils.EmptyStr},
				{"i", "min_int", int64(math.MinInt64 + 1), utils.EmptyStr},
				{"i", "min_int", int64(math.MinInt64 + 2), utils.EmptyStr},
			},
		},
		{
			name: "Increment non-integer values",
			commands: []struct {
				op          string
				key         string
				val         interface{}
				expectedErr string
			}{
				{"s", "float_key", "3.14", utils.EmptyStr},
				{"i", "float_key", nil, "ERR value is not an integer or out of range"},
				{"s", "string_key", "hello", utils.EmptyStr},
				{"i", "string_key", nil, "ERR value is not an integer or out of range"},
				{"s", "bool_key", "true", utils.EmptyStr},
				{"i", "bool_key", nil, "ERR value is not an integer or out of range"},
			},
		},
		{
			name: "Increment non-existent key",
			commands: []struct {
				op          string
				key         string
				val         interface{}
				expectedErr string
			}{
				{"i", "non_existent", int64(1), utils.EmptyStr},
				{"g", "non_existent", int64(1), utils.EmptyStr},
				{"i", "non_existent", int64(2), utils.EmptyStr},
			},
		},
		{
			name: "Increment string representing integers",
			commands: []struct {
				op          string
				key         string
				val         interface{}
				expectedErr string
			}{
				{"s", "str_int1", "42", utils.EmptyStr},
				{"i", "str_int1", int64(43), utils.EmptyStr},
				{"s", "str_int2", "-10", utils.EmptyStr},
				{"i", "str_int2", int64(-9), utils.EmptyStr},
				{"s", "str_int3", "0", utils.EmptyStr},
				{"i", "str_int3", int64(1), utils.EmptyStr},
			},
		},
		{
			name: "Increment with expiry",
			commands: []struct {
				op          string
				key         string
				val         interface{}
				expectedErr string
			}{
				{"se", "expiry_key", int64(0), utils.EmptyStr},
				{"i", "expiry_key", int64(1), utils.EmptyStr},
				{"i", "expiry_key", int64(2), utils.EmptyStr},
				{"w", "expiry_key", nil, utils.EmptyStr},
				{"i", "expiry_key", int64(1), utils.EmptyStr},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up keys before each test case
			keys := []string{"key1", "key2", "max_int", "min_int", "float_key", "string_key", "bool_key",
				"non_existent", "str_int1", "str_int2", "str_int3", "expiry_key"}
			for _, key := range keys {
				FireCommand(conn, fmt.Sprintf("DEL %s", key))
			}

			for _, cmd := range tc.commands {
				switch cmd.op {
				case "s":
					FireCommand(conn, fmt.Sprintf("SET %s %v", cmd.key, cmd.val))
				case "se":
					FireCommand(conn, fmt.Sprintf("SET %s %v EX 1", cmd.key, cmd.val))
				case "i":
					result := FireCommand(conn, fmt.Sprintf("INCR %s", cmd.key))
					switch v := result.(type) {
					case string:
						assert.Equal(t, cmd.expectedErr, v)
					case int64:
						assert.Equal(t, cmd.val, v)
					}
				case "g":
					result := FireCommand(conn, fmt.Sprintf("GET %s", cmd.key))
					assert.Equal(t, cmd.val, result)
				case "w":
					time.Sleep(1100 * time.Millisecond)
				}
			}
		})
	}
}
