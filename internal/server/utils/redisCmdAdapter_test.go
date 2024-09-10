package utils

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TODO: Fix Flaky Tests
func TestParseHTTPRequest(t *testing.T) {
	commands := []struct {
		name         string
		method       string
		url          string
		body         string
		expectedCmd  string
		expectedArgs []string
	}{
		{"Test AUTH command", "POST", "/auth?user=default&password=secret", "", "AUTH", []string{"default", "secret"}},
		{"Test SET command", "POST", "/set?key=k1&value=v1&nx", "", "SET", []string{"k1", "v1", "nx"}},
		{"Test GET command", "POST", "/get?key=k1", "", "GET", []string{"k1"}},
		{"Test MSET command", "POST", "/mset?key1=v1&key2=v2", "", "MSET", []string{"v1", "v2"}},
		{"Test JSON.SET command", "POST", "/json.set", `{"key": "k1", "path": ".", "json": {"field": "value"}}`, "JSON.SET", []string{"key", "k1", "path", ".", "json", `{"field":"value"}`}},
		{"Test JSON.GET command", "POST", "/json.get?key=k1", "", "JSON.GET", []string{"k1"}},
		{"Test DEL command with multiple keys", "POST", "/del?key=k1&key=k2", "", "DEL", []string{"k1", "k2"}},
		{"Test EXPIRE command", "POST", "/expire?key=k1&seconds=100", "", "EXPIRE", []string{"k1", "100"}},
		{"Test HSET command with JSON body", "POST", "/hset", `{"key": "hashkey", "field": "f1", "value": "v1"}`, "HSET", []string{"key", "hashkey", "field", "f1", "value", "v1"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.url, nil)
			}

			redisCmd, err := ParseHTTPRequest(req)
			assert.NoError(t, err)

			expectedCmd := &cmd.RedisCmd{
				Cmd:  tc.expectedCmd,
				Args: tc.expectedArgs,
			}
			assert.Equal(t, expectedCmd, redisCmd, "The parsed Redis command should match the expected command")
		})
	}
}
