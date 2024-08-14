package tests

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/dicedb/dice/internal/constants"
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
				{"s", "key1", 3, constants.EmptyStr},
				{"d", "key1", 2, constants.EmptyStr},
				{"d", "key1", 1, constants.EmptyStr},
				{"d", "key2", -1, constants.EmptyStr},
				{"g", "key1", 1, constants.EmptyStr},
				{"g", "key2", -1, constants.EmptyStr},
				{"s", "key3", math.MinInt64 + 1, constants.EmptyStr},
				{"d", "key3", math.MinInt64, constants.EmptyStr},
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

func TestDECRBY(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	type SetCommand struct {
		key string
		val int64
	}

	type DecrByCommand struct {
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
		decrByCommands []DecrByCommand
		getCommands    []GetCommand
	}{
		{
			name: "Decrement multiple keys",
			setCommands: []SetCommand{
				{"key1", 3},
				{"key3", math.MinInt64 + 1},
			},
			decrByCommands: []DecrByCommand{
				{"key1", int64(2), 1, constants.EmptyStr},
				{"key1", int64(1), 0, constants.EmptyStr},
				{"key4", int64(1), -1, constants.EmptyStr},
				{"key3", int64(1), math.MinInt64, constants.EmptyStr},
				{"key3", int64(math.MinInt64), 0, "ERR value is out of range"},
				{"key5", "abc", 0, "ERR value is not an integer or out of range"},
			},
			getCommands: []GetCommand{
				{"key1", 0},
				{"key4", -1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.setCommands {
				fireCommand(conn, fmt.Sprintf("SET %s %d", cmd.key, cmd.val))
			}

			for _, cmd := range tc.decrByCommands {
				var result any
				switch v := cmd.decrValue.(type) {
				case int64:
					result = fireCommand(conn, fmt.Sprintf("DECRBY %s %d", cmd.key, v))
				case string:
					result = fireCommand(conn, fmt.Sprintf("DECRBY %s %s", cmd.key, v))
				}
				switch v := result.(type) {
				case string:
					assert.Equal(t, cmd.expectedErr, v)
				case int64:
					assert.Equal(t, cmd.expectedVal, v)
				}
			}

			for _, cmd := range tc.getCommands {
				result := fireCommand(conn, fmt.Sprintf("GET %s", cmd.key))
				assert.Equal(t, strconv.FormatInt(cmd.expectedVal, 10), result)
			}
		})
	}
}
