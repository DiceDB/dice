package websocket

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestQWatch(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect interface{}
	}{
		{
			name:   "Wrong number of arguments",
			cmds:   []string{"Q.WATCH "},
			expect: "ERR wrong number of arguments for 'q.watch' command",
		},
		{
			name:   "Invalid query",
			cmds:   []string{"Q.WATCH \"SELECT \""},
			expect: "error parsing SQL statement: syntax error at position 8",
		},
		// TODO - once following query is registered, websocket will also attempt sending updates
		// while keys are set for other tests in this package
		// Add unregister test case to handle this scenario once qunwatch support is added
		{
			name:   "Successful register",
			cmds:   []string{`Q.WATCH "SELECT $key, $value WHERE $key like 'test-key?'"`},
			expect: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-key?'", []interface{}{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				if _, ok := tc.expect.(string); ok {
					// compare strings
					assert.Equal(t, tc.expect, result, "Value mismatch for cmd %s", cmd)
				} else {
					// compare lists
					assert.ElementsMatch(t, tc.expect, result, "Value mismatch for cmd %s", cmd)
				}
			}
		})
	}
}

func TestQWatchMultipleClients(t *testing.T) {
	numOfClients := 10
	exec := NewWebsocketCommandExecutor()
	clients := []*websocket.Conn{}

	for i := 0; i < numOfClients; i++ {
		client := exec.ConnectToServer()
		clients = append(clients, client)
	}

	tc := struct {
		cmd    string
		expect interface{}
	}{
		cmd:    `Q.WATCH "SELECT $key, $value WHERE $key like 'key'"`,
		expect: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'key'", []interface{}{}},
	}

	for i := 0; i < numOfClients; i++ {
		conn := clients[i]

		// subscribe query
		resp, err := exec.FireCommandAndReadResponse(conn, tc.cmd)
		assert.Nil(t, err)
		assert.ElementsMatch(t, tc.expect, resp, "Value mismatch for cmd %s", tc.cmd)
	}

	// update key
	resp, err := exec.FireCommandAndReadResponse(clients[0], "SET key 1")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp, "Value mismatch for cmd SET key 1")

	// read and validate query updates
	for i := 0; i < numOfClients; i++ {
		resp, err := exec.ReadResponse(clients[i], tc.cmd)
		assert.Nil(t, err)
		assert.Equal(t, []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'key'", []interface{}{[]interface{}{"key", float64(1)}}}, resp, "Value mismatch for reading query update message")
	}
}
