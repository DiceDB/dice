package websocket

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/logger"
	testifyAssert "github.com/stretchr/testify/assert"

	"gotest.tools/v3/assert"
)

type TestCase struct {
	name     string
	commands []WebsocketCommand
	expected []interface{}
}

func RunTestServer() {
	var wg sync.WaitGroup
	logger := logger.New(logger.Opts{WithTimestamp: false})
	opts := TestServerOptions{
		Port:   8380,
		Logger: logger,
	}
	RunWebsocketServer(context.Background(), &wg, opts)
}

func TestSet(t *testing.T) {
	// RunTestServer()
	exec := NewWebsocketCommandExecutor()

	testCases := []TestCase{
		{
			name: "Set and Get Simple Value",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v"}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", "v"},
		},
		{
			name: "Set and Get Integer Value",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": 123456789}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", string("1.23456789e+08")},
		},
		{
			name: "Overwrite Existing Key",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v1"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": 5}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", "OK", float64(5)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// delete existing key
			_, err := exec.FireCommand(WebsocketCommand{
				Message: map[string]interface{}{"command": "del", "key": "k"},
			})
			testifyAssert.NoError(t, err)

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				testifyAssert.NoError(t, err)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithOptions(t *testing.T) {
	// RunTestServer()
	exec := NewWebsocketCommandExecutor()
	expiryTime := strconv.FormatInt(time.Now().Add(1*time.Minute).UnixMilli(), 10)

	testCases := []TestCase{
		{
			name: "Set with EX option",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "ex": 3}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
				{Message: map[string]interface{}{"command": "sleep", "key": 3}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "Set with PX option",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "px": 2000}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
				{Message: map[string]interface{}{"command": "sleep", "key": 3}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "Set with EX and PX option",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "ex": 2, "px": 2000}},
			},
			expected: []interface{}{"ERR syntax error"},
		},
		{
			name: "XX on non-existing key",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "del", "key": "k"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "xx": true}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{float64(0), "(nil)", "(nil)"},
		},
		{
			name: "NX on non-existing key",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "del", "key": "k"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "nx": true}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{float64(0), "OK", "v"},
		},
		{
			name: "NX on existing key",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "del", "key": "k"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "nx": true}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "nx": true}},
			},
			expected: []interface{}{float64(0), "OK", "v", "(nil)"},
		},
		{
			name: "PXAT option",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "pxat": expiryTime}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", "v"},
		},
		{
			name: "PXAT option with delete",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k1", "value": "v1", "pxat": expiryTime}},
				{Message: map[string]interface{}{"command": "get", "key": "k1"}},
				{Message: map[string]interface{}{"command": "sleep", "key": 4}},
				{Message: map[string]interface{}{"command": "del", "key": "k1"}},
			},
			expected: []interface{}{"OK", "v1", "OK", float64(1)},
		},
		{
			name: "PXAT option with invalid unix time ms",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k2", "value": "v2", "pxat": "123123"}},
				{Message: map[string]interface{}{"command": "get", "key": "k2"}},
			},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name: "XX on existing key",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v1"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v2", "xx": true}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", "OK", "v2"},
		},
		{
			name: "Multiple XX operations",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v1"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v2", "xx": true}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v3", "xx": true}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", "OK", "OK", "v3"},
		},
		{
			name: "EX option",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "ex": 1}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
				{Message: map[string]interface{}{"command": "sleep", "key": 2}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "XX option",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "xx": true, "ex": 1}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
				{Message: map[string]interface{}{"command": "sleep", "key": 2}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "xx": true, "ex": 1}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
			},
			expected: []interface{}{"(nil)", "(nil)", "OK", "(nil)", "(nil)", "(nil)"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(WebsocketCommand{Message: map[string]interface{}{"command": "del", "key": "k"}})
			exec.FireCommand(WebsocketCommand{Message: map[string]interface{}{"command": "del", "key": "k1"}})
			exec.FireCommand(WebsocketCommand{Message: map[string]interface{}{"command": "del", "key": "k2"}})
			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithExat(t *testing.T) {
	// RunTestServer()
	exec := NewWebsocketCommandExecutor()
	Etime := strconv.FormatInt(time.Now().Unix()+5, 10)
	BadTime := "123123"

	testCases := []TestCase{
		{
			name: "SET with EXAT",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "del", "key": "k"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "exat": Etime}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
				{Message: map[string]interface{}{"command": "TTL", "key": "k"}},
			},
			expected: []interface{}{float64(0), "OK", "v", float64(4)},
		},
		{
			name: "SET with invalid EXAT expires key immediately",
			commands: []WebsocketCommand{
				{Message: map[string]interface{}{"command": "del", "key": "k"}},
				{Message: map[string]interface{}{"command": "set", "key": "k", "value": "v", "exat": BadTime}},
				{Message: map[string]interface{}{"command": "get", "key": "k"}},
				{Message: map[string]interface{}{"command": "TTL", "key": "k"}},
			},
			expected: []interface{}{float64(0), "OK", "(nil)", float64(-2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure key is deleted before the test
			exec.FireCommand(WebsocketCommand{
				Message: map[string]interface{}{"command": "del", "key": "k"},
			})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				assert.NilError(t, err)
				if val, ok := cmd.Message["command"]; ok && val == "TTL" {
					assert.Assert(t, result.(float64) <= tc.expected[i].(float64))
				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}
