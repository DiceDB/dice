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
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Route != expected.Route {
		t.Errorf("expected route %q, got %q", expected.Route, result.Route)
	}
	fmt.Println("result", result, "expected", expected)
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
	expected = "MSET key1 value1 key2 value2 key3 "
	result, err = EncodeCommand(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
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
