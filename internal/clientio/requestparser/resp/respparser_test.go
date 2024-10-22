package respparser

import (
	"reflect"
	"testing"

	"github.com/dicedb/dice/internal/cmd"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []*cmd.DiceDBCmd
		wantErr bool
	}{
		{
			name:  "Simple SET command",
			input: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "SET", Args: []string{"key", "value"}},
			},
		},
		{
			name:  "GET command",
			input: "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "GET", Args: []string{"key"}},
			},
		},
		{
			name:  "Multiple commands",
			input: "*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "PING", Args: []string{"PONG"}},
				{Cmd: "SET", Args: []string{"key", "value"}},
			},
		},
		{
			name:  "Command with integer argument",
			input: "*3\r\n$6\r\nEXPIRE\r\n$3\r\nkey\r\n:60\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "EXPIRE", Args: []string{"key", "60"}},
			},
		},
		{
			name:    "Invalid command (not an array)",
			input:   "NOT AN ARRAY\r\n",
			wantErr: true,
		},
		{
			name:    "Empty command",
			input:   "*0\r\n",
			wantErr: true,
		},
		{
			name:  "Command with null bulk string argument",
			input: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$-1\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "SET", Args: []string{"key", "(nil)"}},
			},
		},
		{
			name:  "Command with Simple String argument",
			input: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n+OK\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "SET", Args: []string{"key", "OK"}},
			},
		},
		{
			name:  "Command with Error argument",
			input: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n-ERR Invalid argument\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "SET", Args: []string{"key", "ERR Invalid argument"}},
			},
		},
		{
			name:  "Command with mixed argument types",
			input: "*5\r\n$4\r\nMSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n:1000\r\n+OK\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "MSET", Args: []string{"key", "value", "1000", "OK"}},
			},
		},
		{
			name:    "Invalid array length",
			input:   "*-2\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			wantErr: true,
		},
		{
			name:    "Incomplete command",
			input:   "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n",
			wantErr: true,
		},
		{
			name:  "Command with empty bulk string",
			input: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$0\r\n\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "SET", Args: []string{"key", ""}},
			},
		},
		{
			name:    "Invalid bulk string length",
			input:   "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$-2\r\nvalue\r\n",
			wantErr: true,
		},
		{
			name:    "Non-integer bulk string length",
			input:   "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$abc\r\nvalue\r\n",
			wantErr: true,
		},
		{
			name:  "Large bulk string",
			input: "*2\r\n$4\r\nECHO\r\n$1000\r\n" + string(make([]byte, 1000)) + "\r\n",
			want: []*cmd.DiceDBCmd{
				{Cmd: "ECHO", Args: []string{string(make([]byte, 1000))}},
			},
		},
		{
			name:    "Incomplete CRLF",
			input:   "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			got, err := p.Parse([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseSimpleString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Valid simple string", "+OK\r\n", "OK", false},
		{"Empty simple string", "+\r\n", "", false},
		{"Simple string with spaces", "+Hello World\r\n", "Hello World", false},
		{"Incomplete simple string", "+OK", "", true},
		{"Missing CR", "+OK\n", "", true},
		{"Missing LF", "+OK\r", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{data: []byte(tt.input), pos: 0}
			got, err := p.parseSimpleString()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSimpleString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSimpleString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Valid error", "-Error message\r\n", "Error message", false},
		{"Empty error", "-\r\n", "", false},
		{"Error with spaces", "-ERR unknown command\r\n", "ERR unknown command", false},
		{"Incomplete error", "-Error", "", true},
		{"Missing CR", "-Error\n", "", true},
		{"Missing LF", "-Error\r", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{data: []byte(tt.input), pos: 0}
			got, err := p.parseError()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseInteger(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"Valid positive integer", ":1000\r\n", 1000, false},
		{"Valid negative integer", ":-1000\r\n", -1000, false},
		{"Zero", ":0\r\n", 0, false},
		{"Large integer", ":9223372036854775807\r\n", 9223372036854775807, false},
		{"Invalid integer (float)", ":3.14\r\n", 0, true},
		{"Invalid integer (text)", ":abc\r\n", 0, true},
		{"Incomplete integer", ":123", 0, true},
		{"Missing CR", ":123\n", 0, true},
		{"Missing LF", ":123\r", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{data: []byte(tt.input), pos: 0}
			got, err := p.parseInteger()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseBulkString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"Valid bulk string", "$5\r\nhello\r\n", "hello", false},
		{"Empty bulk string", "$0\r\n\r\n", "", false},
		{"Null bulk string", "$-1\r\n", "(nil)", false},
		{"Bulk string with spaces", "$11\r\nhello world\r\n", "hello world", false},
		{"Invalid length (negative)", "$-2\r\nhello\r\n", "", true},
		{"Invalid length (non-numeric)", "$abc\r\nhello\r\n", "", true},
		{"Incomplete bulk string", "$5\r\nhell", "", true},
		{"Missing CR", "$5\r\nhello\n", "", true},
		{"Missing LF", "$5\r\nhello\r", "", true},
		{"Length mismatch", "$4\r\nhello\r\n", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{data: []byte(tt.input), pos: 0}
			got, err := p.parseBulkString()
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBulkString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseBulkString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseArray(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "Valid array",
			input: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			want:  []string{"hello", "world"},
		},
		{
			name:    "Empty array",
			input:   "*0\r\n",
			wantErr: true,
		},
		{
			name:    "Null array",
			input:   "*-1\r\n",
			wantErr: true,
		},
		{
			name:  "Array with mixed types",
			input: "*3\r\n:1\r\n$5\r\nhello\r\n+world\r\n",
			want:  []string{"1", "hello", "world"},
		},
		{
			name:    "Invalid array length",
			input:   "*-2\r\n",
			wantErr: true,
		},
		{
			name:    "Non-numeric array length",
			input:   "*abc\r\n",
			wantErr: true,
		},
		{
			name:    "Array length mismatch (too few elements)",
			input:   "*3\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			wantErr: true,
		},
		{
			name:  "Array length mismatch (too many elements)",
			input: "*1\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			want:  []string{"hello"}, // Truncated parsing
		},
		{
			name:    "Incomplete array",
			input:   "*2\r\n$5\r\nhello\r\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{data: []byte(tt.input), pos: 0}
			got, err := p.parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
