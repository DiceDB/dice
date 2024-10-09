package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDataCmd(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	testCases := []struct {
		name          string
		commands      []HTTPCommand
		expected      []interface{}
		assert_type   []string
		errorExpected bool
	}{
		{
			name: "SADD simple value",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "array"},
			expected:    []interface{}{float64(1), []any{string("bar")}},
		},
		{
			name: "SADD multiple values",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), []any{string("bar"), string("baz")}},
		},
		{
			name: "SADD duplicate values",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(0), []any{string("bar")}},
		},
		{
			name: "SADD wrong key value type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
			},
			assert_type: []string{"equal", "equal"},
			expected:    []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "SADD multiple add and multiple kind of values",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "1"}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), float64(1), []any{string("bar"), string("baz"), string("1")}},
		},
		{
			name: "SADD & SCARD",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SCARD", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal", "equal"},
			expected:    []interface{}{float64(1), float64(1), float64(2)},
		},
		{
			name: "SADD & SCARD with non-existing key",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SCARD", Body: map[string]interface{}{"key": "bar"}},
			},
			assert_type: []string{"equal", "equal", "equal"},
			expected:    []interface{}{float64(1), float64(1), float64(0)},
		},
		{
			name: "SADD & SCARD with wrong key type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SCARD", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal"},
			expected:    []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "SADD & SMEMBERS",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), []any{string("bar"), string("baz")}},
		},
		{
			name: "SADD & SMEMBERS with non-existing key",
			commands: []HTTPCommand{
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal"},
			expected:    []interface{}{[]any{}},
		},
		{
			name: "SADD & SMEMBERS with wrong key type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal"},
			expected:    []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "SADD & SREM",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SREM", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), float64(1), []any{string("baz")}},
		},
		{
			name: "SADD & SREM with non-existing key",
			commands: []HTTPCommand{
				{Command: "SREM", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
			},
			assert_type: []string{"equal"},
			expected:    []interface{}{float64(0)},
		},
		{
			name: "SADD & SREM with wrong key type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SREM", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
			},
			assert_type: []string{"equal", "equal"},
			expected:    []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "SADD & SREM with non-existing value",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "values": []interface{}{"bar", "baz", "bax"}}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
				{Command: "SREM", Body: map[string]interface{}{"key": "foo", "value": "bat"}},
				{Command: "SMEMBERS", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "array", "equal", "array"},
			expected:    []interface{}{float64(3), []any{string("bar"), string("baz"), string("bax")}, float64(0), []any{string("bar"), string("baz"), string("bax")}},
		},
		{
			name: "SADD & SDIFF",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo2", "value": "baz"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo2", "value": "bax"}},
				{Command: "SDIFF", Body: map[string]interface{}{"values": []interface{}{"foo", "foo2"}}},
			},
			assert_type: []string{"equal", "equal", "equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), float64(1), float64(1), []any{string("bar")}},
		},
		{
			name: "SADD & SDIFF with non-existing subsequent key",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SDIFF", Body: map[string]interface{}{"values": []interface{}{"foo", "foo2"}}},
			},
			assert_type: []string{"equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), []any{string("bar"), string("baz")}},
		},
		{
			name: "SADD & SDIFF with wrong key type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SDIFF", Body: map[string]interface{}{"values": []interface{}{"foo", "foo2"}}},
			},
			assert_type: []string{"equal", "equal"},
			expected:    []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "SADD & SDIFF with subsequent key of wrong type",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SET", Body: map[string]interface{}{"key": "foo2", "value": "bar"}},
				{Command: "SDIFF", Body: map[string]interface{}{"values": []interface{}{"foo", "foo2"}}},
			},
			assert_type: []string{"equal", "equal", "equal", "equal"},
			expected:    []interface{}{float64(1), float64(1), "OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "SADD & SDIFF with non-existing first key",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SDIFF", Body: map[string]interface{}{"key1": "foo2", "key2": "foo"}},
			},
			assert_type: []string{"equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), []any{}},
		},
		{
			name: "SADD & SDIFF with one key",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SDIFF", Body: map[string]interface{}{"key": "foo"}},
			},
			assert_type: []string{"equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), []any{string("bar"), string("baz")}},
		},
		{
			name: "SADD & SINTER",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo2", "value": "baz"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo2", "value": "bax"}},
				{Command: "SINTER", Body: map[string]interface{}{"values": []interface{}{"foo", "foo2"}}},
			},
			assert_type: []string{"equal", "equal", "equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), float64(1), float64(1), []any{string("baz")}},
		},
		{
			name: "SADD & SINTER with non-existing subsequent key",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SINTER", Body: map[string]interface{}{"values": []interface{}{"foo", "foo2"}}},
			},
			assert_type: []string{"equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), []any{}},
		},
		{
			name: "SADD & SINTER with wrong key type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SINTER", Body: map[string]interface{}{"values": []interface{}{"foo", "foo2"}}},
			},
			assert_type: []string{"equal", "equal"},
			expected:    []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "SADD & SINTER with subsequent key of wrong type",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SET", Body: map[string]interface{}{"key": "foo2", "value": "bar"}},
				{Command: "SINTER", Body: map[string]interface{}{"values": []interface{}{"foo", "foo2"}}},
			},
			assert_type: []string{"equal", "equal", "equal", "equal"},
			expected:    []interface{}{float64(1), float64(1), "OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "SADD & SINTER with single key",
			commands: []HTTPCommand{
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "bar"}},
				{Command: "SADD", Body: map[string]interface{}{"key": "foo", "value": "baz"}},
				{Command: "SINTER", Body: map[string]interface{}{"values": []interface{}{"foo"}}},
			},
			assert_type: []string{"equal", "equal", "array"},
			expected:    []interface{}{float64(1), float64(1), []any{string("bar"), string("baz")}},
		},
	}

	defer exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body:    map[string]interface{}{"key": "foo"},
	})
	defer exec.FireCommand(HTTPCommand{
		Command: "DEL",
		Body:    map[string]interface{}{"key": "foo2"},
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo"},
			})
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "foo2"},
			})

			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				switch tc.assert_type[i] {
				case "array":
					assert.ElementsMatch(t, tc.expected[i], result)
				default:
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}
