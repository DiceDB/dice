package http

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

var qWatchQuery = "SELECT $key, $value WHERE $key LIKE \"match:100:*\" AND $value > 10 ORDER BY $value DESC LIMIT 3"

func TestQWatch(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []TestCase{
		{
			name: "Q.WATCH Register Bad Request",
			commands: []HTTPCommand{
				{Command: "Q.WATCH", Body: map[string]interface{}{}},
			},
			expected: []interface{}{
				[]interface{}{},
			},
			errorExpected: true,
		},
		{
			name: "Q.WATCH Register",
			commands: []HTTPCommand{
				{Command: "Q.WATCH", Body: map[string]interface{}{"query": qWatchQuery}},
			},
			expected: []interface{}{
				map[string]interface{}{
					"cmd":   "q.watch",
					"query": "SELECT $key, $value WHERE $key like 'match:100:*' and $value > 10 ORDER BY $value desc LIMIT 3",
					"data":  []interface{}{},
				},
			},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec.FireCommand(HTTPCommand{
				Command: "DEL",
				Body:    map[string]interface{}{"key": "match:100:user"},
			})

			for i, cmd := range tc.commands {
				result, err := exec.FireCommand(cmd)
				if tc.errorExpected {
					assert.NotNil(t, err)
				} else {
					assert.Equal(t, tc.expected[i], result)
				}
			}
		})
	}
}

func TestQwatchWithSSE(t *testing.T) {
	exec := NewHTTPCommandExecutor()
	const key = "match:100:user:3"
	const val = 15

	qwatchResponseReceived := make(chan struct{}, 2)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		resp, err := http.Post("http://localhost:8083/q.watch", "application/json",
			bytes.NewBuffer([]byte(`{
				"query": "SELECT $key, $value WHERE $key like 'match:100:*' and $value > 10 ORDER BY $value desc LIMIT 3"
			}`)))
		if err != nil {
			t.Errorf("Failed to start QWATCH: %v", err)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				t.Errorf("Failed to close Resp body: %v", err)
			}
		}(resp.Body)

		decoder := json.NewDecoder(resp.Body)
		expectedResponses := []interface{}{
			map[string]interface{}{
				"cmd":   "q.watch",
				"query": "SELECT $key, $value WHERE $key like 'match:100:*' and $value > 10 ORDER BY $value desc LIMIT 3",
				"data":  []interface{}{},
			},
			map[string]interface{}{
				"cmd":   "q.watch",
				"query": "SELECT $key, $value WHERE $key like 'match:100:*' and $value > 10 ORDER BY $value desc LIMIT 3",
				"data":  []interface{}{[]interface{}{key, float64(val)}},
			},
		}

		for responseCount := 0; responseCount < 2; responseCount++ {
			var result interface{}
			if err := decoder.Decode(&result); err != nil {
				if err == io.EOF {
					break
				}
				t.Errorf("Error reading SSE response: %v", err)
				return
			}

			assert.Equal(t, result, expectedResponses[responseCount])
			qwatchResponseReceived <- struct{}{}
		}
	}()

	time.Sleep(2 * time.Second)

	setCmd := HTTPCommand{
		Command: "SET",
		Body:    map[string]interface{}{"key": key, "value": val},
	}
	result, _ := exec.FireCommand(setCmd)
	assert.Equal(t, result, "OK")

	for i := 0; i < 2; i++ {
		select {
		case <-qwatchResponseReceived:
		case <-time.After(10 * time.Second):
			t.Fatalf("Timed out waiting for SSE response")
		}
	}

	wg.Wait()
}
