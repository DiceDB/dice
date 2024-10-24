package resp

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

var deqRandGenerator *rand.Rand
var deqRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_!@#$%^&*()-=+[]\\;':,.<>/?~.|")

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

func deqTestInit() {
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
}

func TestLPush(t *testing.T) {
	deqTestInit()
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func TestRPush(t *testing.T) {
	deqTestInit()
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
	}{
		{
			name:   "RPUSH",
			cmds:   []string{"LPUSH k v", "LPUSH k v1 1 v2 2", "LPUSH k 3 3 3 v3 v3 v3"},
			expect: []any{int64(1), int64(5), int64(11)},
		},
		{
			name:   "RPUSH normal values",
			cmds:   []string{"LPUSH k " + strings.Join(deqNormalValues, " ")},
			expect: []any{int64(25)},
		},
		{
			name:   "RPUSH edge values",
			cmds:   []string{"LPUSH k " + strings.Join(deqEdgeValues, " ")},
			expect: []any{int64(42)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func TestLPushLPop(t *testing.T) {
	deqTestInit()
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func TestLPushRPop(t *testing.T) {
	deqTestInit()
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func TestRPushLPop(t *testing.T) {
	deqTestInit()
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func TestRPushRPop(t *testing.T) {
	deqTestInit()
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func TestLRPushLRPop(t *testing.T) {
	deqTestInit()
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func TestLLEN(t *testing.T) {
	deqTestInit()
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	deqCleanUp(conn, "k")
}

func deqCleanUp(conn net.Conn, key string) {
	for {
		result := FireCommand(conn, "LPOP "+key)
		if result == "(nil)" {
			break
		}
	}
}
