package dicedbadapter

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestFlattenArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      interface{}
		expected  string
		expectErr bool
	}{
		{
			name:     "Single string",
			args:     "SET",
			expected: "SET",
		},
		{
			name:     "Flat list of strings",
			args:     []interface{}{"SET", "key", "value"},
			expected: "SET key value",
		},
		{
			name: "Simple nested list",
			args: []interface{}{
				"SET",
				"key",
				"value",
				[]interface{}{"EX", "100"},
				[]interface{}{"NX"},
			},
			expected: "SET key value EX 100 NX",
		},
		{
			name: "Three levels of nesting",
			args: []interface{}{
				"SET",
				[]interface{}{"key", []interface{}{"value", "EX"}},
				"100",
			},
			expected: "SET key value EX 100",
		},
		{
			name: "Complex nested list",
			args: []interface{}{
				"SET",
				"key",
				"value",
				[]interface{}{"PX", "200"},
				[]interface{}{"XX", "KEEPTTL"},
			},
			expected: "SET key value PX 200 XX KEEPTTL",
		},
		{
			name:     "Number in the list",
			args:     []interface{}{"SET", "key", 100},
			expected: "SET key 100",
		},
		{
			name:      "Completely invalid type",
			args:      123, // root level is an integer, which is unexpected
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := flattenStringArgs(tt.args)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected result %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeCommandSet(t *testing.T) {
	cmd := DiceDBCommand{
		Route: "/set",
		Body: map[string]interface{}{
			"key":   "myKey",
			"value": "myValue",
			"ex":    "60",
		},
	}
	expected := "SET myKey myValue ex 60"
	result, err := EncodeCommand(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestDecodeCommandSetWithKeetTTL(t *testing.T) {
	commandStr := "SET myKey myValue keepTTL"
	expected := DiceDBCommand{
		Route: "/set",
		Body: map[string]interface{}{
			"key":     "myKey",
			"value":   "myValue",
			"keepttl": "keepttl",
		},
	}

	result, err := DecodeCommand(commandStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Route != expected.Route {
		t.Errorf("expected route %q, got %q", expected.Route, result.Route)
	}
	fmt.Println(result)
	for k, v := range expected.Body {
		if result.Body[k] != v {
			t.Errorf("expected body %q for key %q, got %q", v, k, result.Body[k])
		}
	}
}

func TestDecodeCommandSetWithEX(t *testing.T) {
	commandStr := "SET myKey myValue ex 60"
	expected := DiceDBCommand{
		Route: "/set",
		Body: map[string]interface{}{
			"key":   "myKey",
			"value": "myValue",
			"ex":    "60",
		},
	}

	result, err := DecodeCommand(commandStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Route != expected.Route {
		t.Errorf("expected route %q, got %q", expected.Route, result.Route)
	}
	fmt.Println(result)
	for k, v := range expected.Body {
		if result.Body[k] != v {
			t.Errorf("expected body %q for key %q, got %q", v, k, result.Body[k])
		}
	}
}

func TestEncodeCommandGet(t *testing.T) {
	cmd := DiceDBCommand{
		Route: "/get",
		Body: map[string]interface{}{
			"key": "myKey",
		},
	}
	expected := "GET myKey"
	result, err := EncodeCommand(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestDecodeCommandGet(t *testing.T) {
	commandStr := "GET myKey"
	expected := DiceDBCommand{
		Route: "/get",
		Body: map[string]interface{}{
			"key": "myKey",
		},
	}

	result, err := DecodeCommand(commandStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Route != expected.Route {
		t.Errorf("expected route %q, got %q", expected.Route, result.Route)
	}
	for k, v := range expected.Body {
		if result.Body[k] != v {
			t.Errorf("expected body %q for key %q, got %q", v, k, result.Body[k])
		}
	}
}

func TestEncodeCommandMGET(t *testing.T) {
	cmd := DiceDBCommand{
		Route: "/mget",
		Body: map[string]interface{}{
			"keys": []interface{}{"key1", "key2", "key3"},
		},
	}
	expected := "MGET key1 key2 key3"
	result, err := EncodeCommand(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestDecodeCommandMGET(t *testing.T) {
	commandStr := "MGET key1 key2 key3"
	expected := DiceDBCommand{
		Route: "/mget",
		Body: map[string]interface{}{
			"keys": []interface{}{"key1", "key2", "key3"},
		},
	}

	result, err := DecodeCommand(commandStr)
	assert.Nil(t, err)
	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)
}

func TestEncodeCommandMSET(t *testing.T) {
	cmd := DiceDBCommand{
		Route: "/mset",
		Body: map[string]interface{}{
			"key-values": []interface{}{[]interface{}{"key1", "value1"}, []interface{}{"key2", "value2"}, []interface{}{"key3", "value3"}},
		},
	}
	expected := "MSET key1 value1 key2 value2 key3 value3"
	result, err := EncodeCommand(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	// Test with missing values
	cmd = DiceDBCommand{
		Route: "/mset",
		Body: map[string]interface{}{
			"key-values": []interface{}{[]interface{}{"key1", "value1"}, []interface{}{"key2", "value2"}, []interface{}{"key3", ""}},
		},
	}
	expected = "MSET key1 value1 key2 value2 key3"
	result, err = EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestDecodeCommandMSET(t *testing.T) {
	commandStr := "MSET key1 value1 key2 value2 key3 value3"
	expected := DiceDBCommand{
		Route: "/mset",
		Body: map[string]interface{}{
			"key-values": []interface{}{[]interface{}{"key1", "value1"}, []interface{}{"key2", "value2"}, []interface{}{"key3", "value3"}},
		},
	}

	result, err := DecodeCommand(commandStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)

	// Test with missing values
	commandStr = "MSET key1 value1 key2 value2 key3 "
	expected = DiceDBCommand{
		Route: "/mset",
		Body: map[string]interface{}{
			"key-values": []interface{}{[]interface{}{"key1", "value1"}, []interface{}{"key2", "value2"}, []interface{}{"key3", ""}},
		},
	}

	result, err = DecodeCommand(commandStr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)
}

func TestEncodeCommandDEL(t *testing.T) {
	cmd := DiceDBCommand{
		Route: "/del",
		Body: map[string]interface{}{
			"keys": []interface{}{"key1", "key2", "key3"},
		},
	}
	expected := "DEL key1 key2 key3"
	result, err := EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestDecodeCommandDEL(t *testing.T) {
	commandStr := "DEL key1 key2 key3"
	expected := DiceDBCommand{
		Route: "/del",
		Body: map[string]interface{}{
			"keys": []interface{}{"key1", "key2", "key3"},
		},
	}

	result, err := DecodeCommand(commandStr)
	assert.Nil(t, err)
	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)
}

func TestEncodeCommandBitop(t *testing.T) {
	cmd := DiceDBCommand{
		Route: "/bitop",
		Body: map[string]interface{}{
			"operation": "AND",
			"destkey":   "destkey",
			"keys":      []interface{}{"key1", "key2", "key3"},
		},
	}
	expected := "BITOP AND destkey key1 key2 key3"
	result, err := EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)

	// NOT operation
	cmd = DiceDBCommand{
		Route: "/bitop",
		Body: map[string]interface{}{
			"operation": "NOT",
			"destkey":   "destkey",
			"keys":      []interface{}{"key1"},
		},
	}
	expected = "BITOP NOT destkey key1"
	result, err = EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)

	//Missing destkey
	cmd = DiceDBCommand{
		Route: "/bitop",
		Body: map[string]interface{}{
			"operation": "NOT",
			"keys":      []interface{}{"key1"},
		},
	}
	expected = "BITOP NOT key1"
	result, err = EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestDecodeCommandBitop(t *testing.T) {
	commandStr := "BITOP AND destkey key1 key2 key3"
	expected := DiceDBCommand{
		Route: "/bitop",
		Body: map[string]interface{}{
			"operation": "AND",
			"destkey":   "destkey",
			"keys":      []interface{}{"key1", "key2", "key3"},
		},
	}

	result, err := DecodeCommand(commandStr)
	assert.Nil(t, err)
	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)

	// NOT operation
	commandStr = "BITOP NOT destkey key1"
	expected = DiceDBCommand{
		Route: "/bitop",
		Body: map[string]interface{}{
			"operation": "NOT",
			"destkey":   "destkey",
			"keys":      []interface{}{"key1"},
		},
	}

	result, err = DecodeCommand(commandStr)
	assert.Nil(t, err)
	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)

	//Missing destkey
	commandStr = "BITOP NOT key1"
	expected = DiceDBCommand{
		Route: "/bitop",
		Body: map[string]interface{}{
			"operation": "NOT",
			"destkey":   "key1",
			"keys":      []interface{}{},
		},
	}

	result, err = DecodeCommand(commandStr)
	assert.Nil(t, err)
	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)
}

func TestEncodeCommandBitfield(t *testing.T) {
	cmd := DiceDBCommand{
		Route: "/bitfield",
		Body: map[string]interface{}{
			"key": "myKey",
			"subcommands": []map[string]interface{}{
				{
					"subcommand": "get",
					"encoding":   "u4",
					"offset":     "0",
				},
				{
					"subcommand": "set",
					"encoding":   "u4",
					"offset":     "0",
					"value":      "1",
				},
			},
		},
	}
	expected := "BITFIELD myKey get u4 0 set u4 0 1"
	result, err := EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestDecodeCommandBitfield(t *testing.T) {
	commandStr := "BITFIELD myKey GET u4 0 SET u4 0 1"
	expected := DiceDBCommand{
		Route: "/bitfield",
		Body: map[string]interface{}{
			"key": "myKey",
			"subcommands": []map[string]interface{}{
				{
					"subcommand": "get",
					"encoding":   "u4",
					"offset":     "0",
				},
				{
					"subcommand": "set",
					"encoding":   "u4",
					"offset":     "0",
					"value":      "1",
				},
			},
		},
	}

	result, err := DecodeCommand(commandStr)
	assert.Nil(t, err)
	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)
}

func TestEncodeCommandZadd(t *testing.T) {
	cmd := DiceDBCommand{
		Route: "/zadd",
		Body: map[string]interface{}{
			"key": "myKey",
			"score-members": []interface{}{
				[]interface{}{"1", "one"},
				[]interface{}{"2", "two"},
			},
		},
	}
	expected := "ZADD myKey 1 one 2 two"
	result, err := EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)

	// zadd with NX and CH flags
	cmd = DiceDBCommand{
		Route: "/zadd",
		Body: map[string]interface{}{
			"key": "myKey",
			"score-members": []interface{}{
				[]interface{}{"1", "one"},
				[]interface{}{"2", "two"},
			},
			"nx": "nx",
			"ch": "ch",
		},
	}
	expected = "ZADD myKey nx ch 1 one 2 two"
	result, err = EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)

	// zadd with extra flags
	cmd = DiceDBCommand{
		Route: "/zadd",
		Body: map[string]interface{}{
			"key": "myKey",
			"score-members": []interface{}{
				[]interface{}{"1", "one"},
				[]interface{}{"2", "two"},
			},
			"xx":   "xx",
			"incr": "incr",
			"ch":   "ch",
			"lt":   "lt",
			"gt":   "gt",
		},
	}
	expected = "ZADD myKey xx gt lt ch incr 1 one 2 two"
	result, err = EncodeCommand(cmd)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestDecodeCommandZadd(t *testing.T) {
	commandStr := "ZADD myKey nx ch 1 one 2 two"
	expected := DiceDBCommand{
		Route: "/zadd",
		Body: map[string]interface{}{
			"key": "myKey",
			"score-members": []interface{}{
				[]interface{}{"1", "one"},
				[]interface{}{"2", "two"},
			},
			"nx": "nx",
			"ch": "ch",
		},
	}

	result, err := DecodeCommand(commandStr)
	assert.Nil(t, err)
	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)

	// zadd with extra flags
	commandStr = "ZADD myKey xx gt lt ch incr 1 one 2 two"
	expected = DiceDBCommand{
		Route: "/zadd",
		Body: map[string]interface{}{
			"key": "myKey",
			"score-members": []interface{}{
				[]interface{}{"1", "one"},
				[]interface{}{"2", "two"},
			},
			"xx":   "xx",
			"incr": "incr",
			"ch":   "ch",
			"lt":   "lt",
			"gt":   "gt",
		},
	}

	result, err = DecodeCommand(commandStr)
	assert.Nil(t, err)
	assert.Equal(t, expected.Route, result.Route)
	assert.Equal(t, expected.Body, result.Body)
}
