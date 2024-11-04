package http

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	testCases := []TestCase{
		{
			name: "LPUSH",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v"}}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v1", 1, "v2", 2}}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{3, 3, 3, "v3", "v3", "v3"}}},
			},
			expected: []interface{}{float64(1), float64(5), float64(11)},
		},
		{
			name: "LPUSH normal values",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": deqNormalValues}},
			},
			expected: []interface{}{float64(25)},
		},
		{
			name: "LPUSH edge values",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": deqEdgeValues}},
			},
			expected: []interface{}{float64(42)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k"}}})
}

func TestRPush(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	testCases := []TestCase{
		{
			name: "RPUSH",
			commands: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v"}}},
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v1", 1, "v2", 2}}},
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{3, 3, 3, "v3", "v3", "v3"}}},
			},
			expected: []interface{}{float64(1), float64(5), float64(11)},
		},
		{
			name: "RPUSH normal values",
			commands: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": deqNormalValues}},
			},
			expected: []interface{}{float64(25)},
		},
		{
			name: "RPUSH edge values",
			commands: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": deqEdgeValues}},
			},
			expected: []interface{}{float64(42)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k"}}})
}

func TestLPushLPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	getPops := func(values []string) []HTTPCommand {
		pops := make([]HTTPCommand, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = HTTPCommand{
				Command: "LPOP",
				Body:    map[string]interface{}{"key": "k"},
			}
		}
		return pops
	}
	getPopExpects := func(values []string) []interface{} {
		expects := make([]interface{}, len(values))
		for i := 0; i < len(values); i++ {
			expects[i] = values[len(values)-1-i]
		}
		return expects
	}

	testCases := []TestCase{
		{
			name: "LPUSH LPOP",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v1", 1}}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(2), "1", "v1", nil},
		},
		{
			name: "LPUSH LPOP normal values",
			commands: append(
				[]HTTPCommand{
					{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": deqNormalValues}},
				},
				getPops(deqNormalValues)...,
			),
			expected: append(
				append(
					[]interface{}{float64(14)},
					getPopExpects(deqNormalValues)...),
				nil,
			),
		},
		{
			name: "LPUSH LPOP edge values",
			commands: append(
				[]HTTPCommand{
					{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": deqEdgeValues}},
				},
				getPops(deqEdgeValues)...,
			),
			expected: append(
				append(
					[]interface{}{float64(17)},
					getPopExpects(deqEdgeValues)...),
				nil,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k"}}})
}

func TestLPushRPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	getPops := func(values []string) []HTTPCommand {
		pops := make([]HTTPCommand, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = HTTPCommand{
				Command: "RPOP",
				Body:    map[string]interface{}{"key": "k"},
			}
		}
		return pops
	}
	getPopExpects := func(values []string) []interface{} {
		expects := make([]interface{}, len(values))
		for i := 0; i < len(values); i++ {
			expects[i] = values[i]
		}
		return expects
	}

	testCases := []TestCase{
		{
			name: "LPUSH RPOP",
			commands: []HTTPCommand{
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v1", 1}}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(2), "v1", "1", nil},
		},
		{
			name: "LPUSH RPOP normal values",
			commands: append(
				[]HTTPCommand{
					{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": deqNormalValues}},
				},
				getPops(deqNormalValues)...,
			),
			expected: append(
				append(
					[]interface{}{float64(14)},
					getPopExpects(deqNormalValues)...,
				),
				nil,
			),
		},
		{
			name: "LPUSH RPOP edge values",
			commands: append(
				[]HTTPCommand{
					{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": deqEdgeValues}},
				},
				getPops(deqEdgeValues)...,
			),
			expected: append(
				append(
					[]interface{}{float64(17)},
					getPopExpects(deqEdgeValues)...),
				nil,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k"}}})
}

func TestRPushLPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	getPops := func(values []string) []HTTPCommand {
		pops := make([]HTTPCommand, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = HTTPCommand{
				Command: "LPOP",
				Body:    map[string]interface{}{"key": "k"},
			}
		}
		return pops
	}
	getPopExpects := func(values []string) []interface{} {
		expects := make([]interface{}, len(values))
		for i := 0; i < len(values); i++ {
			expects[i] = values[i]
		}
		return expects
	}

	testCases := []TestCase{
		{
			name: "RPUSH LPOP",
			commands: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v1", 1}}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(2), "v1", "1", nil},
		},
		{
			name: "RPUSH LPOP normal values",
			commands: append(
				[]HTTPCommand{
					{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": deqNormalValues}},
				},
				getPops(deqNormalValues)...,
			),
			expected: append(
				append(
					[]interface{}{float64(14)},
					getPopExpects(deqNormalValues)...,
				),
				nil,
			),
		},
		{
			name: "RPUSH LPOP edge values",
			commands: append(
				[]HTTPCommand{
					{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": deqEdgeValues}},
				},
				getPops(deqEdgeValues)...,
			),
			expected: append(
				append(
					[]interface{}{float64(17)},
					getPopExpects(deqEdgeValues)...),
				nil,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k"}}})
}

func TestRPushRPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	getPops := func(values []string) []HTTPCommand {
		pops := make([]HTTPCommand, len(values)+1)
		for i := 0; i < len(values)+1; i++ {
			pops[i] = HTTPCommand{
				Command: "RPOP",
				Body:    map[string]interface{}{"key": "k"},
			}
		}
		return pops
	}
	getPopExpects := func(values []string) []interface{} {
		expects := make([]interface{}, len(values))
		for i := 0; i < len(values); i++ {
			expects[i] = values[len(values)-1-i]
		}
		return expects
	}

	testCases := []TestCase{
		{
			name: "RPUSH RPOP",
			commands: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v1", 1}}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{float64(2), "1", "v1", nil},
		},
		{
			name: "RPUSH RPOP normal values",
			commands: append(
				[]HTTPCommand{
					{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": deqNormalValues}},
				},
				getPops(deqNormalValues)...,
			),
			expected: append(
				append(
					[]interface{}{float64(14)},
					getPopExpects(deqNormalValues)...,
				),
				nil,
			),
		},
		{
			name: "RPUSH RPOP edge values",
			commands: append(
				[]HTTPCommand{
					{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": deqEdgeValues}},
				},
				getPops(deqEdgeValues)...,
			),
			expected: append(
				append(
					[]interface{}{float64(17)},
					getPopExpects(deqEdgeValues)...),
				nil,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k"}}})
}

func TestLRPushLRPop(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	testCases := []TestCase{
		{
			name: "L/RPush L/RPop",
			commands: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v1000", 1000}}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v2000", 2000}}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v6"}}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []interface{}{
				float64(2), float64(4),
				"1000", "v1000", "2000",
				float64(2),
				"v2000", "v6", nil, nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k"}}})
}

func TestLLEN(t *testing.T) {
	deqTestInit()
	exec := NewHTTPCommandExecutor()
	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})

	testCases := []TestCase{
		{
			name: "LLEN",
			commands: []HTTPCommand{
				{Command: "RPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v1000", 1000}}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v2000", 2000}}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPUSH", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v6"}}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "RPOP", Body: map[string]interface{}{"key": "k"}},
				{Command: "LLEN", Body: map[string]interface{}{"key": "k"}},
			},
			expected: []any{
				float64(2), float64(4), float64(4),
				"1000", float64(3), "v1000", "2000", float64(1),
				float64(2), float64(2),
				"v2000", float64(1), "v6", nil, nil, float64(0),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}

	exec.FireCommand(HTTPCommand{Command: "DEL", Body: map[string]interface{}{"keys": [...]string{"k"}}})
}
