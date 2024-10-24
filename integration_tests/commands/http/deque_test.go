package http

import (
	"fmt"
	"math/rand"
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
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name   string
		cmds   []HTTPCommand
		expect []any
	}{
		{
			name: "LPUSH",
			cmds: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]any{"key": "k", "value": "v"}},
				{Command: "LPUSH", Body: map[string]any{"key": "k", "values": []string{"v1", "1", "v2", "2"}}},
				{Command: "LPUSH", Body: map[string]any{"key": "k", "values": []string{"3", "3", "3", "v3", "v3", "v3"}}},
			},
			expect: []any{float64(1), float64(5), float64(11)},
		},
		{
			name: "LPUSH normal values",
			cmds: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]any{"key": "k", "values": deqNormalValues}},
			},
			expect: []any{float64(25)},
		},
		{
			name: "LPUSH edge values",
			cmds: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]any{"key": "k", "values": deqEdgeValues}},
			},
			expect: []any{float64(42)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}
func TestRPush(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name   string
		cmds   []HTTPCommand
		expect []any
	}{
		{
			name: "RPUSH",
			cmds: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]any{"key": "k", "value": "v"}},
				{Command: "RPUSH", Body: map[string]any{"key": "k", "values": []string{"v1", "1", "v2", "2"}}},
				{Command: "RPUSH", Body: map[string]any{"key": "k", "values": []string{"3", "3", "3", "v3", "v3", "v3"}}},
			},
			expect: []any{float64(1), float64(5), float64(11)},
		},
		{
			name: "RPUSH normal values",
			cmds: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]any{"key": "k", "values": deqNormalValues}},
			},
			expect: []any{float64(25)},
		},
		{
			name: "RPUSH edge values",
			cmds: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]any{"key": "k", "values": deqEdgeValues}},
			},
			expect: []any{float64(42)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestLPushLPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()

	getPops := func(values []string) []HTTPCommand {
		pops := make([]HTTPCommand, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = HTTPCommand{Command: "LPOP", Body: map[string]interface{}{"key": "k"}}
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
		cmds   []HTTPCommand
		expect []any
	}{
		{
			name: "LPUSH LPOP",
			cmds: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []string{"v1", "1"}}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expect: []any{float64(2), "1", "v1", nil},
		},
		{
			name: "LPUSH LPOP normal values",
			cmds: append(
				[]HTTPCommand{
					{Command: "LPUSH", Body: map[string]any{"key": "k", "values": deqNormalValues}},
				},
				getPops(deqNormalValues)...,
			),
			expect: append(append([]any{float64(14)}, getPopExpects(deqNormalValues)...), nil),
		},
		{
			name: "LPUSH LPOP edge values",
			cmds: append(
				[]HTTPCommand{
					{Command: "LPUSH", Body: map[string]any{"key": "k", "values": deqEdgeValues}},
				},
				getPops(deqEdgeValues)...,
			),
			expect: append(append([]any{float64(17)}, getPopExpects(deqEdgeValues)...), nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestLPushRPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()

	getPops := func(values []string) []HTTPCommand {
		pops := make([]HTTPCommand, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = HTTPCommand{Command: "RPOP", Body: map[string]interface{}{"key": "k"}}
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
		cmds   []HTTPCommand
		expect []any
	}{
		{
			name: "LPUSH RPOP",
			cmds: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []string{"v1", "1"}}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expect: []any{float64(2), "v1", "1", nil},
		},
		{
			name: "LPUSH RPOP normal values",
			cmds: append(
				[]HTTPCommand{
					{Command: "LPUSH", Body: map[string]any{"key": "k", "values": deqNormalValues}},
				},
				getPops(deqNormalValues)...,
			),
			expect: append(append([]any{float64(14)}, getPopExpects(deqNormalValues)...), nil),
		},
		{
			name: "LPUSH RPOP edge values",
			cmds: append(
				[]HTTPCommand{
					{Command: "LPUSH", Body: map[string]any{"key": "k", "values": deqEdgeValues}},
				},
				getPops(deqEdgeValues)...,
			),
			expect: append(append([]any{float64(17)}, getPopExpects(deqEdgeValues)...), nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestRPushLPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()

	getPops := func(values []string) []HTTPCommand {
		pops := make([]HTTPCommand, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = HTTPCommand{Command: "LPOP", Body: map[string]interface{}{"key": "k"}}
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
		cmds   []HTTPCommand
		expect []any
	}{
		{
			name: "RPUSH LPOP",
			cmds: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []string{"v1", "1"}}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expect: []any{float64(2), "v1", "1", nil},
		},
		{
			name: "RPUSH LPOP normal values",
			cmds: append(
				[]HTTPCommand{
					{Command: "RPUSH", Body: map[string]any{"key": "k", "values": deqNormalValues}},
				},
				getPops(deqNormalValues)...,
			),
			expect: append(append([]any{float64(14)}, getPopExpects(deqNormalValues)...), nil),
		},
		{
			name: "RPUSH LPOP edge values",
			cmds: append(
				[]HTTPCommand{
					{Command: "RPUSH", Body: map[string]any{"key": "k", "values": deqEdgeValues}},
				},
				getPops(deqEdgeValues)...,
			),
			expect: append(append([]any{float64(17)}, getPopExpects(deqEdgeValues)...), nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestRPushRPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()

	getPops := func(values []string) []HTTPCommand {
		pops := make([]HTTPCommand, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = HTTPCommand{Command: "RPOP", Body: map[string]interface{}{"key": "k"}}
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
		cmds   []HTTPCommand
		expect []any
	}{
		{
			name: "RPUSH RPOP",
			cmds: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []string{"v1", "1"}}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expect: []any{float64(2), "1", "v1", nil},
		},
		{
			name: "RPUSH RPOP normal values",
			cmds: append(
				[]HTTPCommand{
					{Command: "RPUSH", Body: map[string]any{"key": "k", "values": deqNormalValues}},
				},
				getPops(deqNormalValues)...,
			),
			expect: append(append([]any{float64(14)}, getPopExpects(deqNormalValues)...), nil),
		},
		{
			name: "RPUSH RPOP edge values",
			cmds: append(
				[]HTTPCommand{
					{Command: "RPUSH", Body: map[string]any{"key": "k", "values": deqEdgeValues}},
				},
				getPops(deqEdgeValues)...,
			),
			expect: append(append([]any{float64(17)}, getPopExpects(deqEdgeValues)...), nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestLRPushLRPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name   string
		cmds   []HTTPCommand
		expect []any
	}{
		{
			name: "L/RPush L/RPop",
			cmds: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []string{"v1000", "1000"}}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []string{"v2000", "2000"}}},

				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},

				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "value": "v6"}},

				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expect: []any{
				float64(2), float64(4),
				"1000", "v1000", "2000",
				float64(2),
				"v2000", "v6", nil, nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}

func TestLLEN(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name   string
		cmds   []HTTPCommand
		expect []any
	}{
		{
			name: "L/RPush L/RPop",
			cmds: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []string{"v1000", "1000"}}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []string{"v2000", "2000"}}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},

				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},

				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "value": "v6"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},

				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
			},
			expect: []any{
				float64(2), float64(4), float64(4),
				"1000", float64(3), "v1000", "2000", float64(1),
				float64(2), float64(2),
				"v2000", float64(1), "v6", nil, nil, float64(0),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"key": "k"}})
}
