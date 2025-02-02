// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

var deqRandGenerator *rand.Rand
var deqRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var (
	deqNormalValues []string
	deqEdgeValues   []string
)

func deqRandStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = deqRunes[deqRandGenerator.Intn(len(deqRunes))]
	}
	return string(b)
}

func deqTestInit() (deqNormalValues, deqEdgeValues []string) {
	randSeed := time.Now().UnixNano()
	deqRandGenerator = rand.New(rand.NewSource(randSeed))
	fmt.Printf("rand seed: %v", randSeed)
	deqNormalValues = []string{
		deqRandStr(10),               // 6 bit string
		deqRandStr(256),              // 12 bit string
		deqRandStr((1 << 13) - 1000), // 32 bit string
		"28",                         // 7 bit uint
		"2024",                       // + 13 bit int
		"-2024",                      // - 13 bit int
		"15384",                      // + 16 bit int
		"-15384",                     // - 16 bit int
		"4193301",                    // + 24 bit int
		"-4193301",                   // - 24 bit int
		"1073731765",                 // + 32 bit int
		"-1073731765",                // - 32 bit int
		"4611686018427287903",        // + 64 bit int
		"-4611686018427287903",       // - 64 bit int
	}
	deqEdgeValues = []string{
		deqRandStr(1),             // min 6 bit string
		deqRandStr((1 << 6) - 1),  // max 6 bit string
		deqRandStr(1 << 6),        // min 12 bit string
		deqRandStr((1 << 12) - 1), // max 12 bit string
		deqRandStr(1 << 12),       // min 32 bit string
		// randStr((1 << 32) - 1),   // max 32 bit string, maybe too huge to test.

		"0",                    // min 7 bit uint
		"127",                  // max 7 bit uint
		"-4096",                // min 13 bit int
		"4095",                 // max 13 bit int
		"-32768",               // min 16 bit int
		"32767",                // max 16 bit int
		"-8388608",             // min 24 bit int
		"8388607",              // max 24 bit int
		"-2147483648",          // min 32 bit int
		"2147483647",           // max 32 bit int
		"-9223372036854775808", // min 64 bit int
		"9223372036854775807",  // max 64 bit int
	}
	return deqNormalValues, deqEdgeValues
}

func TestLPush(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "LPUSH",
			cmds:   []string{"LPUSH k v", "LPUSH k v1 1 v2 2", "LPUSH k 3 3 3 v3 v3 v3"},
			expect: []any{float64(1), float64(5), float64(11)},
		},
		{
			name:   "LPUSH normal values",
			cmds:   []string{"LPUSH k " + strings.Join(deqNormalValues, " ")},
			expect: []any{float64(25)},
		},
		{
			name:   "LPUSH edge values",
			cmds:   []string{"LPUSH k " + strings.Join(deqEdgeValues, " ")},
			expect: []any{float64(42)},
		},
	}

	conn := exec.ConnectToServer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	DeleteKey(t, conn, exec, "k")
}

func TestRPush(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "RPUSH",
			cmds:   []string{"RPUSH k v", "RPUSH k v1 1 v2 2", "RPUSH k 3 3 3 v3 v3 v3"},
			expect: []any{float64(1), float64(5), float64(11)},
		},
		{
			name:   "RPUSH normal values",
			cmds:   []string{"RPUSH k " + strings.Join(deqNormalValues, " ")},
			expect: []any{float64(25)},
		},
		{
			name:   "RPUSH edge values",
			cmds:   []string{"RPUSH k " + strings.Join(deqEdgeValues, " ")},
			expect: []any{float64(42)},
		},
	}

	conn := exec.ConnectToServer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	DeleteKey(t, conn, exec, "k")
}

func TestLPushLPop(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	exec := NewWebsocketCommandExecutor()

	getPops := func(values []string) []string {
		pops := make([]string, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = "LPOP k"
		}
		return pops
	}
	getPopExpects := func(values []string) []any {
		expects := make([]any, len(values))
		for i := 0; i < len(values); i++ {
			expects[i] = values[len(values)-1-i]
		}
		return expects
	}

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "LPUSH LPOP",
			cmds:   []string{"LPUSH k v1 1", "LPOP k", "LPOP k", "LPOP k"},
			expect: []any{float64(2), "1", "v1", nil},
		},
		{
			name:   "LPUSH LPOP normal values",
			cmds:   append([]string{"LPUSH k " + strings.Join(deqNormalValues, " ")}, getPops(deqNormalValues)...),
			expect: append(append([]any{float64(14)}, getPopExpects(deqNormalValues)...), nil),
		},
		{
			name:   "LPUSH LPOP edge values",
			cmds:   append([]string{"LPUSH k " + strings.Join(deqEdgeValues, " ")}, getPops(deqEdgeValues)...),
			expect: append(append([]any{float64(17)}, getPopExpects(deqEdgeValues)...), nil),
		},
	}

	conn := exec.ConnectToServer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	DeleteKey(t, conn, exec, "k")
}

func TestLPushRPop(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	exec := NewWebsocketCommandExecutor()

	getPops := func(values []string) []string {
		pops := make([]string, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = "RPOP k"
		}
		return pops
	}
	getPopExpects := func(values []string) []any {
		expects := make([]any, len(values))
		for i := 0; i < len(values); i++ {
			expects[i] = values[i]
		}
		return expects
	}

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "LPUSH RPOP",
			cmds:   []string{"LPUSH k v1 1", "RPOP k", "RPOP k", "RPOP k"},
			expect: []any{float64(2), "v1", "1", nil},
		},
		{
			name:   "LPUSH RPOP normal values",
			cmds:   append([]string{"LPUSH k " + strings.Join(deqNormalValues, " ")}, getPops(deqNormalValues)...),
			expect: append(append([]any{float64(14)}, getPopExpects(deqNormalValues)...), nil),
		},
		{
			name:   "LPUSH RPOP edge values",
			cmds:   append([]string{"LPUSH k " + strings.Join(deqEdgeValues, " ")}, getPops(deqEdgeValues)...),
			expect: append(append([]any{float64(17)}, getPopExpects(deqEdgeValues)...), nil),
		},
	}

	conn := exec.ConnectToServer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	DeleteKey(t, conn, exec, "k")
}

func TestRPushLPop(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	exec := NewWebsocketCommandExecutor()

	getPops := func(values []string) []string {
		pops := make([]string, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = "LPOP k"
		}
		return pops
	}
	getPopExpects := func(values []string) []any {
		expects := make([]any, len(values))
		for i := 0; i < len(values); i++ {
			expects[i] = values[i]
		}
		return expects
	}

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "RPUSH LPOP",
			cmds:   []string{"RPUSH k v1 1", "LPOP k", "LPOP k", "LPOP k"},
			expect: []any{float64(2), "v1", "1", nil},
		},
		{
			name:   "RPUSH LPOP normal values",
			cmds:   append([]string{"RPUSH k " + strings.Join(deqNormalValues, " ")}, getPops(deqNormalValues)...),
			expect: append(append([]any{float64(14)}, getPopExpects(deqNormalValues)...), nil),
		},
		{
			name:   "RPUSH LPOP edge values",
			cmds:   append([]string{"RPUSH k " + strings.Join(deqEdgeValues, " ")}, getPops(deqEdgeValues)...),
			expect: append(append([]any{float64(17)}, getPopExpects(deqEdgeValues)...), nil),
		},
	}

	conn := exec.ConnectToServer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	DeleteKey(t, conn, exec, "k")
}

func TestRPushRPop(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	exec := NewWebsocketCommandExecutor()

	getPops := func(values []string) []string {
		pops := make([]string, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = "RPOP k"
		}
		return pops
	}
	getPopExpects := func(values []string) []any {
		expects := make([]any, len(values))
		for i := 0; i < len(values); i++ {
			expects[i] = values[len(values)-1-i]
		}
		return expects
	}

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "RPUSH RPOP",
			cmds:   []string{"RPUSH k v1 1", "RPOP k", "RPOP k", "RPOP k"},
			expect: []any{float64(2), "1", "v1", nil},
		},
		{
			name:   "RPUSH RPOP normal values",
			cmds:   append([]string{"RPUSH k " + strings.Join(deqNormalValues, " ")}, getPops(deqNormalValues)...),
			expect: append(append([]any{float64(14)}, getPopExpects(deqNormalValues)...), nil),
		},
		{
			name:   "RPUSH RPOP edge values",
			cmds:   append([]string{"RPUSH k " + strings.Join(deqEdgeValues, " ")}, getPops(deqEdgeValues)...),
			expect: append(append([]any{float64(17)}, getPopExpects(deqEdgeValues)...), nil),
		},
	}

	conn := exec.ConnectToServer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	DeleteKey(t, conn, exec, "k")
}

func TestLRPushLRPop(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name: "L/RPush L/RPop",
			cmds: []string{
				"RPUSH k v1000 1000", "LPUSH k v2000 2000",
				"RPOP k", "RPOP k", "LPOP k",
				"LPUSH k v6",
				"RPOP k", "LPOP k", "LPOP k", "RPOP k",
			},
			expect: []any{
				float64(2), float64(4),
				"1000", "v1000", "2000",
				float64(2),
				"v2000", "v6", nil, nil,
			},
		},
	}

	conn := exec.ConnectToServer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	DeleteKey(t, conn, exec, "k")
}

func TestLLEN(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name: "L/RPush L/RPop",
			cmds: []string{
				"RPUSH k v1000 1000", "LPUSH k v2000 2000", "LLEN k",
				"RPOP k", "LLEN k", "RPOP k", "LPOP k", "LLEN k",
				"LPUSH k v6", "LLEN k",
				"RPOP k", "LLEN k", "LPOP k", "LPOP k", "RPOP k", "LLEN k",
			},
			expect: []any{
				float64(2), float64(4), float64(4),
				"1000", float64(3), "v1000", "2000", float64(1),
				float64(2), float64(2),
				"v2000", float64(1), "v6", nil, nil, float64(0),
			},
		},
	}

	conn := exec.ConnectToServer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	DeleteKey(t, conn, exec, "k")
}

func TestLPOPCount(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		cleanupKey string
	}{
		{
			name: "LPOP with count argument - valid, invalid, and edge cases",
			commands: []string{
				"RPUSH k v1",
				"RPUSH k v2",
				"RPUSH k v3",
				"RPUSH k v4",
				"LPOP k 2",
				"LPOP k 2",
				"LPOP k -1",
				"LPOP k abc",
				"LLEN k",
			},
			expected: []interface{}{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				[]interface{}{"v1", "v2"},
				[]interface{}{"v3", "v4"},
				"ERR value is not an integer or out of range",
				"ERR value is not an integer or a float",
				float64(0),
			},
			cleanupKey: "k",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()

			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.NilError(t, err)
				assert.DeepEqual(t, tc.expected[i], result)
			}
			DeleteKey(t, conn, exec, tc.cleanupKey)
		})
	}
}
