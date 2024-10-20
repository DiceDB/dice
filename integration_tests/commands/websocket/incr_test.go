package websocket

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestINCR(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "Increment multiple keys",
			cmds:   []string{"SET key1 0", "INCR key1", "INCR key1", "INCR key2", "GET key1", "GET key2"},
			expect: []interface{}{"OK", float64(1), float64(2), float64(1), float64(2), float64(1)},
			delays: []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:   "Increment to and from max int64",
			cmds:   []string{fmt.Sprintf("SET max_int %s", strconv.Itoa(math.MaxInt64-1)), "INCR max_int", "INCR max_int", fmt.Sprintf("SET max_int %s", strconv.Itoa(math.MaxInt64)), "INCR max_int"},
			expect: []interface{}{"OK", float64(math.MaxInt64), "ERR increment or decrement would overflow", "OK", "ERR increment or decrement would overflow"},
			delays: []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name:   "Increment from min int64",
			cmds:   []string{fmt.Sprintf("SET min_int %s", strconv.Itoa(math.MinInt64)), "INCR min_int", "INCR min_int"},
			expect: []interface{}{"OK", float64(math.MinInt64 + 1), float64(math.MinInt64 + 2)},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "Increment non-integer values",
			cmds:   []string{"SET float_key 3.14", "INCR float_key", "SET string_key hello", "INCR string_key", "SET bool_key true", "INCR bool_key"},
			expect: []interface{}{"OK", "ERR value is not an integer or out of range", "OK", "ERR value is not an integer or out of range", "OK", "ERR value is not an integer or out of range"},
			delays: []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:   "Increment non-existent key",
			cmds:   []string{"INCR non_existent", "GET non_existent", "INCR non_existent"},
			expect: []interface{}{float64(1), float64(1), float64(2)},
			delays: []time.Duration{0, 0, 0},
		},
		{
			name:   "Increment string representing integers",
			cmds:   []string{"SET str_int1 42", "INCR str_int1", "SET str_int2 -10", "INCR str_int2", "SET str_int3 0", "INCR str_int3"},
			expect: []interface{}{"OK", float64(43), "OK", float64(-9), "OK", float64(1)},
			delays: []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:   "Increment with expiry",
			cmds:   []string{"SET expiry_key 0 EX 1", "INCR expiry_key", "INCR expiry_key", "INCR expiry_key"},
			expect: []interface{}{"OK", float64(1), float64(2), float64(1)},
			delays: []time.Duration{0, 0, 0, 2 * time.Second},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys := []string{
				"key1", "key2", "max_int", "min_int", "float_key", "string_key", "bool_key",
				"non_existent", "str_int1", "str_int2", "str_int3", "expiry_key",
			}
			for _, key := range keys {
				exec.FireCommandAndReadResponse(conn, fmt.Sprintf("DEL %s", key))
			}

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

func TestINCRBY(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name:   "happy flow",
			cmds:   []string{"SET key 3", "INCRBY key 2", "INCRBY key 1", "GET key"},
			expect: []interface{}{"OK", float64(5), float64(6), float64(6)},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name:   "happy flow with negative increment",
			cmds:   []string{"SET key 100", "INCRBY key -2", "INCRBY key -10", "INCRBY key -88", "INCRBY key -100", "GET key"},
			expect: []interface{}{"OK", float64(98), float64(88), float64(0), float64(-100), float64(-100)},
			delays: []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:   "happy flow with unset key",
			cmds:   []string{"SET key 3", "INCRBY unsetKey 2", "GET key", "GET unsetKey"},
			expect: []interface{}{"OK", float64(2), float64(3), float64(2)},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name:   "edge case with maxInt64",
			cmds:   []string{fmt.Sprintf("SET key %s", strconv.Itoa(math.MaxInt64-1)), "INCRBY key 1", "INCRBY key 1", "GET key"},
			expect: []interface{}{"OK", float64(math.MaxInt64), "ERR increment or decrement would overflow", float64(math.MaxInt64)},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name:   "edge case with negative increment",
			cmds:   []string{fmt.Sprintf("SET key %s", strconv.Itoa(math.MinInt64+1)), "INCRBY key -1", "INCRBY key -1", "GET key"},
			expect: []interface{}{"OK", float64(math.MinInt64), "ERR increment or decrement would overflow", float64(math.MinInt64)},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name:   "edge case with string values",
			cmds:   []string{"SET key 1", "INCRBY stringkey abc"},
			expect: []interface{}{"OK", "ERR value is not an integer or out of range"},
			delays: []time.Duration{0, 0},
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
