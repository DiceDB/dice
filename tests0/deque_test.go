// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/dicedb/dicedb-go"
	"github.com/stretchr/testify/assert"
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
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "LPUSH",
			cmds:   []string{"LPUSH k v", "LPUSH k v1 1 v2 2", "LPUSH k 3 3 3 v3 v3 v3"},
			expect: []any{int64(1), int64(5), int64(11)},
		},
		{
			name:   "LPUSH normal values",
			cmds:   []string{"LPUSH k " + strings.Join(deqNormalValues, " ")},
			expect: []any{int64(25)},
		},
		{
			name:   "LPUSH edge values",
			cmds:   []string{"LPUSH k " + strings.Join(deqEdgeValues, " ")},
			expect: []any{int64(42)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestRPush(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "RPUSH",
			cmds:   []string{"RPUSH k v", "RPUSH k v1 1 v2 2", "RPUSH k 3 3 3 v3 v3 v3"},
			expect: []any{int64(1), int64(5), int64(11)},
		},
		{
			name:   "RPUSH normal values",
			cmds:   []string{"RPUSH k " + strings.Join(deqNormalValues, " ")},
			expect: []any{int64(25)},
		},
		{
			name:   "RPUSH edge values",
			cmds:   []string{"RPUSH k " + strings.Join(deqEdgeValues, " ")},
			expect: []any{int64(42)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestLPushLPop(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	client := getLocalConnection()
	defer client.Close()

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
			expect: []any{int64(2), "1", "v1", "(nil)"},
		},
		{
			name:   "LPUSH LPOP normal values",
			cmds:   append([]string{"LPUSH k " + strings.Join(deqNormalValues, " ")}, getPops(deqNormalValues)...),
			expect: append(append([]any{int64(14)}, getPopExpects(deqNormalValues)...), "(nil)"),
		},
		{
			name:   "LPUSH LPOP edge values",
			cmds:   append([]string{"LPUSH k " + strings.Join(deqEdgeValues, " ")}, getPops(deqEdgeValues)...),
			expect: append(append([]any{int64(17)}, getPopExpects(deqEdgeValues)...), "(nil)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestLPushRPop(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	client := getLocalConnection()
	defer client.Close()

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
			expect: []any{int64(2), "v1", "1", "(nil)"},
		},
		{
			name:   "LPUSH RPOP normal values",
			cmds:   append([]string{"LPUSH k " + strings.Join(deqNormalValues, " ")}, getPops(deqNormalValues)...),
			expect: append(append([]any{int64(14)}, getPopExpects(deqNormalValues)...), "(nil)"),
		},
		{
			name:   "LPUSH RPOP edge values",
			cmds:   append([]string{"LPUSH k " + strings.Join(deqEdgeValues, " ")}, getPops(deqEdgeValues)...),
			expect: append(append([]any{int64(17)}, getPopExpects(deqEdgeValues)...), "(nil)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestRPushLPop(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	client := getLocalConnection()
	defer client.Close()

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
			expect: []any{int64(2), "v1", "1", "(nil)"},
		},
		{
			name:   "RPUSH LPOP normal values",
			cmds:   append([]string{"RPUSH k " + strings.Join(deqNormalValues, " ")}, getPops(deqNormalValues)...),
			expect: append(append([]any{int64(14)}, getPopExpects(deqNormalValues)...), "(nil)"),
		},
		{
			name:   "RPUSH LPOP edge values",
			cmds:   append([]string{"RPUSH k " + strings.Join(deqEdgeValues, " ")}, getPops(deqEdgeValues)...),
			expect: append(append([]any{int64(17)}, getPopExpects(deqEdgeValues)...), "(nil)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestRPushRPop(t *testing.T) {
	deqNormalValues, deqEdgeValues := deqTestInit()
	client := getLocalConnection()
	defer client.Close()

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
			expect: []any{int64(2), "1", "v1", "(nil)"},
		},
		{
			name:   "RPUSH RPOP normal values",
			cmds:   append([]string{"RPUSH k " + strings.Join(deqNormalValues, " ")}, getPops(deqNormalValues)...),
			expect: append(append([]any{int64(14)}, getPopExpects(deqNormalValues)...), "(nil)"),
		},
		{
			name:   "RPUSH RPOP edge values",
			cmds:   append([]string{"RPUSH k " + strings.Join(deqEdgeValues, " ")}, getPops(deqEdgeValues)...),
			expect: append(append([]any{int64(17)}, getPopExpects(deqEdgeValues)...), "(nil)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestLRPushLRPop(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

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
				int64(2), int64(4),
				"1000", "v1000", "2000",
				int64(2),
				"v2000", "v6", "(nil)", "(nil)",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestLLEN(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

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
				int64(2), int64(4), int64(4),
				"1000", int64(3), "v1000", "2000", int64(1),
				int64(2), int64(2),
				"v2000", int64(1), "v6", "(nil)", "(nil)", int64(0),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestLInsert(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "LINSERT before",
			cmds:   []string{"LPUSH k v1 v2 v3 v4", "LINSERT k before v2 e1", "LINSERT k before v1 e2", "LINSERT k before v4 e3", "LRANGE k 0 6"},
			expect: []any{int64(4), int64(5), int64(6), int64(7), []any{"e3", "v4", "v3", "e1", "v2", "e2", "v1"}},
		},
		{
			name:   "LINSERT after",
			cmds:   []string{"LINSERT k after v2 e4", "LINSERT k after v1 e5", "LINSERT k after v4 e6", "LRANGE k 0 10"},
			expect: []any{int64(8), int64(9), int64(10), []any{"e3", "v4", "e6", "v3", "e1", "v2", "e4", "e2", "v1", "e5"}},
		},
		{
			name:   "LINSERT wrong number of args",
			cmds:   []string{"LINSERT k before e1"},
			expect: []any{"-wrong number of arguments for LINSERT"},
		},
		{
			name:   "LINSERT wrong type",
			cmds:   []string{"SET k1 val1", "LINSERT k1 before val1 val2"},
			expect: []any{"OK", "-WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				// assert.DeepEqual(t, tc.expect[i], result)
				assert.EqualValues(t, tc.expect[i], result)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestLRange(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "LRANGE with +ve start stop",
			cmds:   []string{"LPUSH k v1 v2 v3 v4", "LINSERT k before v2 e1", "LINSERT k before v1 e2", "LINSERT k before v4 e3", "LRANGE k 0 6"},
			expect: []any{int64(4), int64(5), int64(6), int64(7), []any{"e3", "v4", "v3", "e1", "v2", "e2", "v1"}},
		},
		{
			name:   "LRANGE with -ve start stop",
			cmds:   []string{"LRANGE k -100 -2"},
			expect: []any{[]any{"e3", "v4", "v3", "e1", "v2", "e2"}},
		},
		{
			name:   "LRANGE wrong number of args",
			cmds:   []string{"LRANGE k -100"},
			expect: []any{"-wrong number of arguments for LRANGE"},
		},
		{
			name:   "LRANGE wrong type",
			cmds:   []string{"SET k1 val1", "LRANGE k1 0 100"},
			expect: []any{"OK", "-WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				// assert.DeepEqual(t, tc.expect[i], result)
				assert.EqualValues(t, tc.expect[i], result)
			}
		})
	}

	deqCleanUp(client, "k")
}

func TestLPOPCount(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
	}{
		{
			name: "LPOP with count argument - valid, invalid, and edge cases",
			cmds: []string{
				"RPUSH k v1 v2 v3 v4",
				"LPOP k 2",
				"LLEN k",
				"LPOP k 0",
				"LLEN k",
				"LPOP k 5",
				"LLEN k",
				"LPOP k -1",
				"LPOP k abc",
				"LLEN k",
			},
			expect: []any{
				int64(4),
				[]interface{}{"v1", "v2"},
				int64(2),
				"(nil)",
				int64(2),
				[]interface{}{"v3", "v4"},
				int64(0),
				"ERR value is not an integer or out of range",
				"ERR value is not an integer or a float",
				int64(0),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result)
			}
		})
	}

	deqCleanUp(client, "k")

}

func deqCleanUp(client *dicedb.Client, key string) {
	for {
		result := client.FireString("LPOP " + key)
		if result.GetVNil() {
			break
		}
	}
}
