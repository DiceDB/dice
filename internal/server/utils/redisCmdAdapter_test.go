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
			name:         "Test JSON.SET command",
			method:       "POST",
			url:          "/json.set",
			body:         `{"key": "k1", "path": ".", "json": {"field": "value"}}`,
			expectedCmd:  "JSON.SET",
			expectedArgs: []string{"k1", ".", `{"field":"value"}`},
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
