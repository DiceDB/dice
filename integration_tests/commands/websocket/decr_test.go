package websocket

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDECR(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "Decrement multiple keys",
			cmds:   []string{"SET key1 3", "DECR key1", "DECR key1", "DECR key2", "GET key1", "GET key2", fmt.Sprintf("SET key3 %s", strconv.Itoa(math.MinInt64+1)), "DECR key3", "DECR key3"},
			expect: []interface{}{"OK", float64(2), float64(1), float64(-1), float64(1), float64(-1), "OK", float64(math.MinInt64), "ERR increment or decrement would overflow"},
			delays: []time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestDECRBY(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "Decrement multiple keys",
			cmds:   []string{"SET key1 3", fmt.Sprintf("SET key3 %s", strconv.Itoa(math.MinInt64+1)), "DECRBY key1 2", "DECRBY key1 1", "DECRBY key4 1", "DECRBY key3 1", fmt.Sprintf("DECRBY key3 %s", strconv.Itoa(math.MinInt64)), "DECRBY key5 abc"},
			expect: []interface{}{"OK", "OK", float64(1), float64(0), float64(-1), float64(math.MinInt64), "ERR increment or decrement would overflow", "ERR value is not an integer or out of range"},
			delays: []time.Duration{0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer exec.FireCommandAndReadResponse(conn, "DEL key unsetKey stringkey")
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
