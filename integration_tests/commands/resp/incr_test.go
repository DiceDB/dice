package resp

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/stretchr/testify/assert"
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
			keys := []string{
				"key1", "key2", "max_int", "min_int", "float_key", "string_key", "bool_key",
				"non_existent", "str_int1", "str_int2", "str_int3", "expiry_key",
			}
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

func TestINCRBY(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	type SetCommand struct {
		key string
		val int64
	}

	type IncrByCommand struct {
		key         string
		decrValue   any
		expectedVal int64
		expectedErr string
	}

	type GetCommand struct {
		key         string
		expectedVal int64
	}

	testCases := []struct {
		name           string
		setCommands    []SetCommand
		incrByCommands []IncrByCommand
		getCommands    []GetCommand
	}{
		{
			name: "happy flow",
			setCommands: []SetCommand{
				{"key", 3},
			},
			incrByCommands: []IncrByCommand{
				{"key", int64(2), 5, utils.EmptyStr},
				{"key", int64(1), 6, utils.EmptyStr},
			},
			getCommands: []GetCommand{
				{"key", 6},
			},
		},
		{
			name: "happy flow with negative increment",
			setCommands: []SetCommand{
				{"key", 100},
			},
			incrByCommands: []IncrByCommand{
				{"key", int64(-2), 98, utils.EmptyStr},
				{"key", int64(-10), 88, utils.EmptyStr},
				{"key", int64(-88), 0, utils.EmptyStr},
				{"key", int64(-100), -100, utils.EmptyStr},
			},
			getCommands: []GetCommand{
				{"key", -100},
			},
		},
		{
			name: "happy flow with unset key",
			setCommands: []SetCommand{
				{"key", 3},
			},
			incrByCommands: []IncrByCommand{
				{"unsetKey", int64(2), 2, utils.EmptyStr},
			},
			getCommands: []GetCommand{
				{"key", 3},
				{"unsetKey", 2},
			},
		},
		{
			name: "edge case with maxInt64",
			setCommands: []SetCommand{
				{"key", math.MaxInt64 - 1},
			},
			incrByCommands: []IncrByCommand{
				{"key", int64(1), math.MaxInt64, utils.EmptyStr},
				{"key", int64(1), 0, "ERR increment or decrement would overflow"},
			},
			getCommands: []GetCommand{
				{"key", math.MaxInt64},
			},
		},
		{
			name: "edge case with negative increment",
			setCommands: []SetCommand{
				{"key", math.MinInt64 + 1},
			},
			incrByCommands: []IncrByCommand{
				{"key", int64(-1), math.MinInt64, utils.EmptyStr},
				{"key", int64(-1), 0, "ERR increment or decrement would overflow"},
			},
			getCommands: []GetCommand{
				{"key", math.MinInt64},
			},
		},
		{
			name: "edge case with string values",
			setCommands: []SetCommand{
				{"key", 1},
			},
			incrByCommands: []IncrByCommand{
				{"stringkey", "abc", 0, "ERR value is not an integer or out of range"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer FireCommand(conn, "DEL key unsetKey stringkey")

			for _, cmd := range tc.setCommands {
				FireCommand(conn, fmt.Sprintf("SET %s %d", cmd.key, cmd.val))
			}

			for _, cmd := range tc.incrByCommands {
				var result any
				switch v := cmd.decrValue.(type) {
				case int64:
					result = FireCommand(conn, fmt.Sprintf("INCRBY %s %d", cmd.key, v))
				case string:
					result = FireCommand(conn, fmt.Sprintf("INCRBY %s %s", cmd.key, v))
				}
				switch v := result.(type) {
				case string:
					assert.Equal(t, cmd.expectedErr, v)
				case int64:
					assert.Equal(t, cmd.expectedVal, v)
				}
			}

			for _, cmd := range tc.getCommands {
				result := FireCommand(conn, fmt.Sprintf("GET %s", cmd.key))
				assert.Equal(t, cmd.expectedVal, result)
			}
		})
	}
}
