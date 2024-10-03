package utils

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/stretchr/testify/assert"
)

func TestParseHTTPRequest(t *testing.T) {
	commands := []struct {
		name         string
		method       string
		url          string
		body         string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "Test SET command with nx flag",
			method:       "POST",
			url:          "/set",
			body:         `{"key": "k1", "value": "v1", "nx": "true"}`,
			expectedCmd:  "SET",
			expectedArgs: []string{"k1", "v1", "nx"},
		},
		{
			name:         "Test GET command",
			method:       "POST",
			url:          "/get",
			body:         `{"key": "k1"}`,
			expectedCmd:  "GET",
			expectedArgs: []string{"k1"},
		},
		{
			name:         "Test DEL command",
			method:       "POST",
			url:          "/del",
			body:         `{"key": "k1"}`,
			expectedCmd:  "DEL",
			expectedArgs: []string{"k1"},
		},
		{
			name:         "Test DEL command with multiple keys",
			method:       "POST",
			url:          "/del",
			body:         `{"keys": ["k1", "k2", "k3"]}`,
			expectedCmd:  "DEL",
			expectedArgs: []string{"k1", "k2", "k3"},
		},
		{
			name:         "Test KEYS command",
			method:       "POST",
			url:          "/keys",
			body:         `{"key": "*name*"}`,
			expectedCmd:  "KEYS",
			expectedArgs: []string{"*name*"},
		},
		{
			name:         "Test MSET command",
			method:       "POST",
			url:          "/mset",
			body:         `{"key_values": {"key1": "v1", "key2": "v2"}}`,
			expectedCmd:  "MSET",
			expectedArgs: []string{"key1", "v1", "key2", "v2"},
		},
		{
			name:         "Test MSET command with options",
			method:       "POST",
			url:          "/mset",
			body:         `{"key_values": {"key1": "v1", "key2": "v2"}, "nx": "true"}`,
			expectedCmd:  "MSET",
			expectedArgs: []string{"key1", "v1", "key2", "v2", "nx"},
		},
		{
			name:         "Test SLEEP command",
			method:       "POST",
			url:          "/sleep",
			body:         `{"key": 10}`,
			expectedCmd:  "SLEEP",
			expectedArgs: []string{"10"},
		},
		{
			name:         "Test PING command",
			method:       "POST",
			url:          "/ping",
			body:         "",
			expectedCmd:  "PING",
			expectedArgs: nil,
		},
		{
			name:         "Test JSON.SET command",
			method:       "POST",
			url:          "/json.set",
			body:         `{"key": "k1", "path": ".", "json": {"field": "value"}}`,
			expectedCmd:  "JSON.SET",
			expectedArgs: []string{"k1", ".", `{"field":"value"}`},
		},
		{
			name:         "Test EXPIRE command",
			method:       "POST",
			url:          "/expire",
			body:         `{"key": "k1", "seconds": "100"}`,
			expectedCmd:  "EXPIRE",
			expectedArgs: []string{"k1", "100"},
		},
		{
			name:         "Test AUTH command",
			method:       "POST",
			url:          "/auth",
			body:         `{"user": "default", "password": "secret"}`,
			expectedCmd:  "AUTH",
			expectedArgs: []string{"default", "secret"},
		},
		{
			name:         "Test JSON.GET command",
			method:       "POST",
			url:          "/json.get",
			body:         `{"key": "k1"}`,
			expectedCmd:  "JSON.GET",
			expectedArgs: []string{"k1"},
		},
		{
			name:         "Test LPUSH command",
			method:       "POST",
			url:          "/lpush",
			body:         `{"key": "k1", "value": "v1"}`,
			expectedCmd:  "LPUSH",
			expectedArgs: []string{"k1", "v1"},
		},
		{
			name:         "Test LPUSH command with multiple items",
			method:       "POST",
			url:          "/lpush",
			body:         `{"key": "k1", "values": ["v1", "v2", "v3"]}`,
			expectedCmd:  "LPUSH",
			expectedArgs: []string{"k1", "v1", "v2", "v3"},
		},
		{
			name:         "Test HSET command with JSON body",
			method:       "POST",
			url:          "/hset",
			body:         `{"key": "hashkey", "field": "f1", "value": "v1"}`,
			expectedCmd:  "HSET",
			expectedArgs: []string{"hashkey", "f1", "v1"},
		},
		{
			name:         "Test JSON.INGEST command",
			method:       "POST",
			url:          "/json.ingest?key_prefix=gmtr_",
			body:         `{"json": {"field": "value"},"path": "$..field"}`,
			expectedCmd:  "JSON.INGEST",
			expectedArgs: []string{"gmtr_", "$..field", `{"field":"value"}`},
		},
		{
			name:         "Test QWATCH command",
			method:       "POST",
			url:          "/qwatch",
			body:         `{"query": "SELECT $key, $value WHERE $key LIKE \"player:*\" AND \"$value.score\" > 10 ORDER BY $value.score DESC LIMIT 5"}`,
			expectedCmd:  "QWATCH",
			expectedArgs: []string{"SELECT $key, $value WHERE $key LIKE \"player:*\" AND \"$value.score\" > 10 ORDER BY $value.score DESC LIMIT 5"},
		},
		{
			name:         "Test JSON.ARRPOP command",
			method:       "POST",
			url:          "/json.arrpop",
			body:         `{"key": "k1", "path": "$", "index": 1}`,
			expectedCmd:  "JSON.ARRPOP",
			expectedArgs: []string{"k1", "$", "1"},
		},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			diceDBCmd, err := ParseHTTPRequest(req)
			assert.NoError(t, err)

			expectedCmd := &cmd.DiceDBCmd{
				Cmd:  tc.expectedCmd,
				Args: tc.expectedArgs,
			}

			// Check command match
			assert.Equal(t, expectedCmd.Cmd, diceDBCmd.Cmd)

			// Check arguments match, regardless of order
			assert.ElementsMatch(t, expectedCmd.Args, diceDBCmd.Args, "The parsed arguments should match the expected arguments, ignoring order")

		})
	}
}

func TestParseWebsocketMessage(t *testing.T) {
	commands := []struct {
		name         string
		message      string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "Test SET command with nx flag",
			message:      "set k1 v1 nx",
			expectedCmd:  "SET",
			expectedArgs: []string{"k1", "v1", "nx"},
		},
		{
			name:         "Test GET command",
			message:      "get k1",
			expectedCmd:  "GET",
			expectedArgs: []string{"k1"},
		},
		{
			name:         "Test JSON.SET command",
			message:      `json.set k1 . {"field":"value"}`,
			expectedCmd:  "JSON.SET",
			expectedArgs: []string{"k1", ".", `{"field":"value"}`},
		},
		{
			name:         "Test JSON.GET command",
			message:      "json.get k1",
			expectedCmd:  "JSON.GET",
			expectedArgs: []string{"k1"},
		},
		{
			name:         "Test HSET command with JSON body",
			message:      "hset hashkey f1 v1",
			expectedCmd:  "HSET",
			expectedArgs: []string{"hashkey", "f1", "v1"},
		},
		{
			name:         "Test JSON.INGEST command with key prefix",
			message:      `json.ingest gmtr_ $..field {"field":"value"}`,
			expectedCmd:  "JSON.INGEST",
			expectedArgs: []string{"gmtr_", "$..field", `{"field":"value"}`},
		},
		{
			name:         "Test JSON.INGEST command without key prefix",
			message:      `json.ingest $..field {"field":"value"}`,
			expectedCmd:  "JSON.INGEST",
			expectedArgs: []string{"", "$..field", `{"field":"value"}`},
		},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			// parse websocket message
			diceDBCmd, err := ParseWebsocketMessage([]byte(tc.message))
			assert.NoError(t, err)

			expectedCmd := &cmd.DiceDBCmd{
				Cmd:  tc.expectedCmd,
				Args: tc.expectedArgs,
			}

			// Check command match
			assert.Equal(t, expectedCmd.Cmd, diceDBCmd.Cmd)

			// Check arguments match, regardless of order
			assert.ElementsMatch(t, expectedCmd.Args, diceDBCmd.Args, "The parsed arguments should match the expected arguments, ignoring order")
		})
	}
}
