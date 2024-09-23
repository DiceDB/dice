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
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			redisCmd, err := ParseHTTPRequest(req)
			assert.NoError(t, err)

			expectedCmd := &cmd.RedisCmd{
				Cmd:  tc.expectedCmd,
				Args: tc.expectedArgs,
			}

			// Check command match
			assert.Equal(t, expectedCmd.Cmd, redisCmd.Cmd)

			// Check arguments match, regardless of order
			assert.ElementsMatch(t, expectedCmd.Args, redisCmd.Args, "The parsed arguments should match the expected arguments, ignoring order")

		})
	}
}
