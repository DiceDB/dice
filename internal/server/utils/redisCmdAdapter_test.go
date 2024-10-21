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
			name:         "Test SET command with value as a map",
			method:       "POST",
			url:          "/set",
			body:         `{"key": "k1", "value": {"subKey": "subValue"}, "nx": "true"}`,
			expectedCmd:  "SET",
			expectedArgs: []string{"k1", `{"subKey":"subValue"}`, "nx"},
		},
		{
			name:         "Test SET command with value as an array",
			method:       "POST",
			url:          "/set",
			body:         `{"key": "k1", "value": ["item1", "item2", "item3"], "nx": "true"}`,
			expectedCmd:  "SET",
			expectedArgs: []string{"k1", `["item1","item2","item3"]`, "nx"},
		},
		{
			name:         "Test SET command with value as a map containing an array",
			method:       "POST",
			url:          "/set",
			body:         `{"key": "k1", "value": {"subKey": ["item1", "item2"]}, "nx": "true"}`,
			expectedCmd:  "SET",
			expectedArgs: []string{"k1", `{"subKey":["item1","item2"]}`, "nx"},
		},
		{
			name:         "Test SET command with value as a deeply nested map",
			method:       "POST",
			url:          "/set",
			body:         `{"key": "k1", "value": {"subKey": {"subSubKey": {"deepKey": "deepValue"}}}, "nx": "true"}`,
			expectedCmd:  "SET",
			expectedArgs: []string{"k1", `{"subKey":{"subSubKey":{"deepKey":"deepValue"}}}`, "nx"},
		},
		{
			name:         "Test SET command with value as an array of maps",
			method:       "POST",
			url:          "/set",
			body:         `{"key": "k1", "value": [{"subKey1": "value1"}, {"subKey2": "value2"}], "nx": "true"}`,
			expectedCmd:  "SET",
			expectedArgs: []string{"k1", `[{"subKey1":"value1"},{"subKey2":"value2"}]`, "nx"},
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
			url:          "/q.watch",
			body:         `{"query": "SELECT $key, $value WHERE $key LIKE \"player:*\" AND \"$value.score\" > 10 ORDER BY $value.score DESC LIMIT 5"}`,
			expectedCmd:  "Q.WATCH",
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
		name          string
		message       string
		expectedCmd   string
		expectedArgs  []string
		expectedError string
	}{
		{
			name:          "SET command with nx flag",
			message:       "set k1 v1 nx",
			expectedCmd:   "SET",
			expectedArgs:  []string{"k1", "v1", "nx"},
			expectedError: "",
		},
		{
			name:          "Test SET command with value as a map",
			message:       `set k0 {"k1":"v1"} nx`,
			expectedCmd:   "SET",
			expectedArgs:  []string{"k0", `{"k1":"v1"}`, "nx"},
			expectedError: "",
		},
		{
			name:          "Test SET command with value as an array",
			message:       `set k1 ["v1","v2","v3"] nx`,
			expectedCmd:   "SET",
			expectedArgs:  []string{"k1", `["v1","v2","v3"]`, "nx"},
			expectedError: "",
		},
		{
			name:          "Test SET command with value as a map containing an array",
			message:       `set k1 {"k2":["v1","v2"]} nx`,
			expectedCmd:   "SET",
			expectedArgs:  []string{"k1", `{"k2":["v1","v2"]}`, "nx"},
			expectedError: "",
		},
		{
			name:          "Test SET command with value as a deeply nested map",
			message:       `set k1 {"k2":{"k3":{"k4":"value"}}} nx`,
			expectedCmd:   "SET",
			expectedArgs:  []string{"k1", `{"k2":{"k3":{"k4":"value"}}}`, "nx"},
			expectedError: "",
		},
		{
			name:          "Test SET command with value as an array of maps",
			message:       `set k0 [{"k1":"v1"},{"k2":"v2"}] nx`,
			expectedCmd:   "SET",
			expectedArgs:  []string{"k0", `[{"k1":"v1"},{"k2":"v2"}]`, "nx"},
			expectedError: "",
		},
		{
			name:          "Test GET command",
			message:       "get k1",
			expectedCmd:   "GET",
			expectedArgs:  []string{"k1"},
			expectedError: "",
		},
		{
			name:          "Test DEL command",
			message:       "del k1",
			expectedCmd:   "DEL",
			expectedArgs:  []string{"k1"},
			expectedError: "",
		},
		{
			name:          "Test DEL command with multiple keys",
			message:       `del k1 k2 k3`,
			expectedCmd:   "DEL",
			expectedArgs:  []string{"k1", "k2", "k3"},
			expectedError: "",
		},
		{
			name:          "Test KEYS command",
			message:       "keys *",
			expectedCmd:   "KEYS",
			expectedArgs:  []string{"*"},
			expectedError: "",
		},
		{
			name:          "Test MSET command",
			message:       "mset k1 v1 k2 v2",
			expectedCmd:   "MSET",
			expectedArgs:  []string{"k1", "v1", "k2", "v2"},
			expectedError: "",
		},
		{
			name:          "Test MSET command with options",
			message:       "mset k1 v1 k2 v2 nx",
			expectedCmd:   "MSET",
			expectedArgs:  []string{"k1", "v1", "k2", "v2", "nx"},
			expectedError: "",
		},
		{
			name:          "Test SLEEP command",
			message:       "sleep 1",
			expectedCmd:   "SLEEP",
			expectedArgs:  []string{"1"},
			expectedError: "",
		},
		{
			name:          "Test PING command",
			message:       "ping",
			expectedCmd:   "PING",
			expectedArgs:  nil,
			expectedError: "",
		},
		{
			name:          "Test EXPIRE command",
			message:       "expire k1 1",
			expectedCmd:   "EXPIRE",
			expectedArgs:  []string{"k1", "1"},
			expectedError: "",
		},
		{
			name:          "Test AUTH command",
			message:       "auth user password",
			expectedCmd:   "AUTH",
			expectedArgs:  []string{"user", "password"},
			expectedError: "",
		},
		{
			name:          "Test LPUSH command",
			message:       "lpush k1 v1",
			expectedCmd:   "LPUSH",
			expectedArgs:  []string{"k1", "v1"},
			expectedError: "",
		},
		{
			name:          "Test LPUSH command with multiple items",
			message:       `lpush k1 v1 v2 v3`,
			expectedCmd:   "LPUSH",
			expectedArgs:  []string{"k1", "v1", "v2", "v3"},
			expectedError: "",
		},
		{
			name:          "Test JSON.ARRPOP command",
			message:       "json.arrpop k1 $ 1",
			expectedCmd:   "JSON.ARRPOP",
			expectedArgs:  []string{"k1", "$", "1"},
			expectedError: "",
		},
		{
			name:          "Test JSON.SET command",
			message:       `json.set k1 . {"field":"value"}`,
			expectedCmd:   "JSON.SET",
			expectedArgs:  []string{"k1", ".", `{"field":"value"}`},
			expectedError: "",
		},
		{
			name:          "Test JSON.GET command",
			message:       "json.get k1",
			expectedCmd:   "JSON.GET",
			expectedArgs:  []string{"k1"},
			expectedError: "",
		},
		{
			name:          "Test HSET command with JSON body",
			message:       "hset hashkey f1 v1",
			expectedCmd:   "HSET",
			expectedArgs:  []string{"hashkey", "f1", "v1"},
			expectedError: "",
		},
		{
			name:          "Test JSON.INGEST command with key prefix",
			message:       `json.ingest gmtr_ $..field {"field":"value"}`,
			expectedCmd:   "JSON.INGEST",
			expectedArgs:  []string{"gmtr_", "$..field", `{"field":"value"}`},
			expectedError: "",
		},
		{
			name:          "Test JSON.INGEST command without key prefix",
			message:       `json.ingest $..field {"field":"value"}`,
			expectedCmd:   "JSON.INGEST",
			expectedArgs:  []string{"", "$..field", `{"field":"value"}`},
			expectedError: "",
		},
		{
			name:          "invalid Q.WATCH no args",
			message:       "q.watch",
			expectedCmd:   "Q.WATCH",
			expectedArgs:  nil,
			expectedError: "",
		},
		{
			name:          "invalid Q.WATCH invalid query",
			message:       `q.watch \"select $key, $value where $key like 'k?'\"`, // backticks will retain escaped characters as it is
			expectedCmd:   "Q.WATCH",
			expectedArgs:  nil,
			expectedError: "error parsing q.watch query: invalid syntax",
		},
		{
			name:          "valid Q.WATCH simple query",
			message:       "q.watch \"select $key, $value where $key like 'k?'\"",
			expectedCmd:   "Q.WATCH",
			expectedArgs:  []string{"select $key, $value where $key like 'k?'"},
			expectedError: "",
		},
		{
			name:          "valid Q.WATCH complex query",
			message:       "q.watch \"SELECT $key, $value WHERE $key LIKE 'player:*' AND '$value.score' > 10 ORDER BY $value.score DESC LIMIT 5\"",
			expectedCmd:   "Q.WATCH",
			expectedArgs:  []string{"SELECT $key, $value WHERE $key LIKE 'player:*' AND '$value.score' > 10 ORDER BY $value.score DESC LIMIT 5"},
			expectedError: "",
		},
		{
			name:          "invalid Q.UNWATCH no args",
			message:       "q.unwatch",
			expectedCmd:   "Q.UNWATCH",
			expectedArgs:  nil,
			expectedError: "",
		},
		{
			name:          "invalid Q.UNWATCH clientID missing",
			message:       "q.unwatch \"select $key, $value where $key like 'k?'\"",
			expectedCmd:   "Q.UNWATCH",
			expectedArgs:  nil,
			expectedError: "error parsing q.unwatch args: clientID or query not found",
		},
		{
			name:          "invalid Q.UNWATCH query missing",
			message:       "q.unwatch 615405144",
			expectedCmd:   "Q.UNWATCH",
			expectedArgs:  nil,
			expectedError: "error parsing q.unwatch args: clientID or query not found",
		},
		{
			name:          "invalid Q.UNWATCH invalid clientID",
			message:       "q.unwatch 61abc5144 \"select $key, $value where $key like 'k?'\"",
			expectedCmd:   "Q.UNWATCH",
			expectedArgs:  nil,
			expectedError: "invalid clientID",
		},
		{
			name:          "invalid Q.UNWATCH negative clientID",
			message:       "q.unwatch -1 \"select $key, $value where $key like 'k?'\"",
			expectedCmd:   "Q.UNWATCH",
			expectedArgs:  nil,
			expectedError: "clientID must be positive",
		},
		{
			name:          "invalid Q.UNWATCH overflowing clientID",
			message:       "q.unwatch 4294967296 \"select $key, $value where $key like 'k?'\"",
			expectedCmd:   "Q.UNWATCH",
			expectedArgs:  nil,
			expectedError: "clientID must be less than 4294967295 (uint32)",
		},
		{
			name:          "invalid Q.UNWATCH invalid query",
			message:       "q.unwatch 615405144 \"select $key, $value where $key like 'k?'", // ending " missing for query
			expectedCmd:   "Q.UNWATCH",
			expectedArgs:  nil,
			expectedError: "error parsing q.unwatch query: invalid syntax",
		},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			// parse websocket message
			diceDBCmd, err := ParseWebsocketMessage([]byte(tc.message))

			// error cases
			if tc.expectedError != "" {
				if err == nil {
					t.Errorf("received nil error but expected error is: %v", tc.expectedError)
				}
				assert.Equal(t, tc.expectedError, err.Error())
				assert.Nil(t, tc.expectedArgs, "received error and not nil args")

				// non error cases
			} else {
				assert.NoError(t, err)
				expectedCmd := &cmd.DiceDBCmd{
					Cmd:  tc.expectedCmd,
					Args: tc.expectedArgs,
				}

				// Check command match
				assert.Equal(t, expectedCmd.Cmd, diceDBCmd.Cmd)

				// Check arguments match, regardless of order
				assert.ElementsMatch(t, expectedCmd.Args, diceDBCmd.Args, "The parsed arguments should match the expected arguments, ignoring order")
			}

		})
	}
}
