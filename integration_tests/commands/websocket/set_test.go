package websocket

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
				{Message: "set k v"},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", "v"},
		},
		{
			name: "Set and Get Integer Value",
			commands: []WebsocketCommand{
				{Message: "set k 123456789"},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", float64(1.23456789e+08)},
		},
		{
			name: "Overwrite Existing Key",
			commands: []WebsocketCommand{
				{Message: "set k v1"},
				{Message: "set k 5"},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", "OK", float64(5)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// delete existing key
			_, err := exec.FireCommand(WebsocketCommand{
				Message: "del k",
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
				{Message: "set k v ex 3"},
				{Message: "get k"},
				{Message: "sleep 3"},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "Set with PX option",
			commands: []WebsocketCommand{
				{Message: "set k v px 2000"},
				{Message: "get k"},
				{Message: "sleep 3"},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "Set with EX and PX option",
			commands: []WebsocketCommand{
				{Message: "set k v ex 2 px 2000"},
			},
			expected: []interface{}{"ERR syntax error"},
		},
		{
			name: "XX on non-existing key",
			commands: []WebsocketCommand{
				{Message: "del k"},
				{Message: "set k v xx true"},
				{Message: "get k"},
			},
			expected: []interface{}{float64(0), "(nil)", "(nil)"},
		},
		{
			name: "NX on non-existing key",
			commands: []WebsocketCommand{
				{Message: "del k"},
				{Message: "set k v nx"},
				{Message: "get k"},
			},
			expected: []interface{}{float64(0), "OK", "v"},
		},
		{
			name: "NX on existing key",
			commands: []WebsocketCommand{
				{Message: "del k"},
				{Message: "set k v nx"},
				{Message: "get k"},
				{Message: "set k v nx"},
			},
			expected: []interface{}{float64(0), "OK", "v", "(nil)"},
		},
		{
			name: "PXAT option",
			commands: []WebsocketCommand{
				{Message: fmt.Sprintf("set k v pxat %v", expiryTime)},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", "v"},
		},
		{
			name: "PXAT option with delete",
			commands: []WebsocketCommand{
				{Message: fmt.Sprintf("set k1 v1 pxat %v", expiryTime)},
				{Message: "get k1"},
				{Message: "sleep 4"},
				{Message: "del k1"},
			},
			expected: []interface{}{"OK", "v1", "OK", float64(1)},
		},
		{
			name: "PXAT option with invalid unix time ms",
			commands: []WebsocketCommand{
				{Message: "set k2 v2 pxat 123123"},
				{Message: "get k2"},
			},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name: "XX on existing key",
			commands: []WebsocketCommand{
				{Message: "set k v2"},
				{Message: "set k v2 xx"},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", "OK", "v2"},
		},
		{
			name: "Multiple XX operations",
			commands: []WebsocketCommand{
				{Message: "set k v1"},
				{Message: "set k v2 xx"},
				{Message: "set k v3 xx"},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", "OK", "OK", "v3"},
		},
		{
			name: "EX option",
			commands: []WebsocketCommand{
				{Message: "set k v ex 1"},
				{Message: "get k"},
				{Message: "sleep 2"},
				{Message: "get k"},
			},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name: "XX option",
			commands: []WebsocketCommand{
				{Message: "set k v xx ex 1"},
				{Message: "get k"},
				{Message: "sleep 2"},
				{Message: "get k"},
				{Message: "set k v xx ex 1"},
				{Message: "get k"},
			},
			expected: []interface{}{"(nil)", "(nil)", "OK", "(nil)", "(nil)", "(nil)"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(WebsocketCommand{Message: "del k"})
			exec.FireCommand(WebsocketCommand{Message: "del k1"})
			exec.FireCommand(WebsocketCommand{Message: "del k2"})
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
				{Message: "del k"},
				{Message: fmt.Sprintf("set k v exat %v", Etime)},
				{Message: "get k"},
				{Message: "ttl k"},
			},
			expected: []interface{}{float64(0), "OK", "v", float64(4)},
		},
		{
			name: "SET with invalid EXAT expires key immediately",
			commands: []WebsocketCommand{
				{Message: "del k"},
				{Message: fmt.Sprintf("set k v exat %v", BadTime)},
				{Message: "get k"},
				{Message: "ttl k"},
			},
			expected: []interface{}{float64(0), "OK", "(nil)", float64(-2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure key is deleted before the test
			exec.FireCommand(WebsocketCommand{
				Message: "del k",
			})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				assert.NilError(t, err)
				command := strings.Split(cmd.Message, "")
				if command[0] == "ttl" {
					assert.Assert(t, result.(float64) <= tc.expected[i].(float64))
				} else {
					assert.DeepEqual(t, tc.expected[i], result)
				}
			}
		})
	}
}
