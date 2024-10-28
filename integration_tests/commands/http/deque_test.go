package http

import (
	"fmt"
	"math/rand"
	// "net"
	// "strings"
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
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}

}
