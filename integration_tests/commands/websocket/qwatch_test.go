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
			cmds:   []string{`Q.WATCH "SELECT $key, $value WHERE $key like 'test-qwatch-key'"`},
			expect: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-qwatch-key'", []interface{}{}},
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

func TestQWatchWithMultipleClients(t *testing.T) {
	numOfClients := 10
	exec := NewWebsocketCommandExecutor()
	clients := []*websocket.Conn{}

	for i := 0; i < numOfClients; i++ {
		client := exec.ConnectToServer()
		defer client.Close()
		clients = append(clients, client)
	}

	tc := struct {
		cmd    string
		expect interface{} // immediate response after firing command
		update interface{} // update received after value change
	}{
		cmd:    `Q.WATCH "SELECT $key, $value WHERE $key like 'test-multiple-clients-key'"`,
		expect: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-multiple-clients-key'", []interface{}{}},
		update: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-multiple-clients-key'", []interface{}{[]interface{}{"test-multiple-clients-key", float64(1)}}},
	}

	for i := 0; i < numOfClients; i++ {
		conn := clients[i]

		// subscribe query
		resp, err := exec.FireCommandAndReadResponse(conn, tc.cmd)
		assert.Nil(t, err)
		assert.ElementsMatch(t, tc.expect, resp, "Value mismatch for cmd %s", tc.cmd)
	}

	// create a fresh client to update keys
	c := exec.ConnectToServer()
	defer c.Close()

	// update key
	resp, err := exec.FireCommandAndReadResponse(c, "SET test-multiple-clients-key 1")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp, "Value mismatch for cmd SET test-multiple-clients-key 1")

	// read and validate query updates
	for i := 0; i < numOfClients; i++ {
		resp, err := exec.ReadResponse(clients[i], tc.cmd)
		assert.Nil(t, err)
		assert.Equal(t, tc.update, resp, "Value mismatch for reading query update message")
	}
}

func TestQWatchWithMultipleClientsAndQueries(t *testing.T) {
	numOfClients := 3
	exec := NewWebsocketCommandExecutor()
	clients := []*websocket.Conn{}

	for i := 0; i < numOfClients; i++ {
		client := exec.ConnectToServer()
		defer client.Close()
		clients = append(clients, client)
	}

	tests := []struct {
		cmd    string
		expect interface{} // immediate response after firing command
		update interface{} // update received after value change
	}{
		{
			cmd:    `Q.WATCH "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key1'"`,
			expect: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key1'", []interface{}{}},
			update: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key1'", []interface{}{[]interface{}{"test-multiple-client-queries-key1", float64(1)}}},
		},
		{
			cmd:    `Q.WATCH "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key2'"`,
			expect: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key2'", []interface{}{}},
			update: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key2'", []interface{}{[]interface{}{"test-multiple-client-queries-key2", float64(2)}}},
		},
		{
			cmd:    `Q.WATCH "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key3'"`,
			expect: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key3'", []interface{}{}},
			update: []interface{}{"q.watch", "SELECT $key, $value WHERE $key like 'test-multiple-client-queries-key3'", []interface{}{[]interface{}{"test-multiple-client-queries-key3", float64(3)}}},
		},
	}

	for i := 0; i < numOfClients; i++ {
		conn := clients[i]

		for j := 0; j < len(tests); j++ {
			// subscribe query
			resp, err := exec.FireCommandAndReadResponse(conn, tests[j].cmd)
			assert.Nil(t, err)
			assert.ElementsMatch(t, tests[j].expect, resp, "Value mismatch for cmd %s", tests[j].cmd)
		}
	}

	// create a fresh client to update keys
	c := exec.ConnectToServer()
	defer c.Close()

	// update keys
	resp, err := exec.FireCommandAndReadResponse(c, "SET test-multiple-client-queries-key1 1")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp, "Value mismatch for cmd SET test-multiple-client-queries-key1 1")

	resp, err = exec.FireCommandAndReadResponse(c, "SET test-multiple-client-queries-key2 2")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp, "Value mismatch for cmd SET test-multiple-client-queries-key2 1")

	resp, err = exec.FireCommandAndReadResponse(c, "SET test-multiple-client-queries-key3 3")
	assert.Nil(t, err)
	assert.Equal(t, "OK", resp, "Value mismatch for cmd SET test-multiple-client-queries-key3 3")

	// prepare update array
	want := []interface{}{}
	for j := 0; j < len(tests); j++ {
		want = append(want, tests[j].update)
	}

	// read and validate query updates
	for i := 0; i < numOfClients; i++ {
		var respArr []interface{}
		for j := 0; j < len(tests); j++ {
			resp, err := exec.ReadResponse(clients[i], tests[0].cmd)
			assert.Nil(t, err)
			// fmt.Println(j, resp)
			respArr = append(respArr, resp)
			// fmt.Println("after append ", respArr)
		}
		// fmt.Println(respArr)
		// fmt.Println(want)
		assert.ElementsMatch(t, want, respArr, "Value mismatch for reading query update message")
	}
}
