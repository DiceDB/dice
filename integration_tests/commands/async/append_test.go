package async

import (
	"math/rand"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func randString(n int, charset string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

func TestAPPEND(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	types := map[string]string{
		"binary": "01",
		"alpha":  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"compr":  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	}

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name: "APPEND After Set and Delete",
			commands: []string{
				"SET key value",
				"APPEND key value",
				"GET key",
				"APPEND key 100",
				"GET key",
				"DEL key",
				"APPEND key value",
				"GET key",
			},
			expected: []interface{}{"OK", int64(10), "valuevalue", int64(13), "valuevalue100", int64(1), int64(5), "value"},
		},
		{
			name: "APPEND to Integer Values",
			commands: []string{
				"DEL key",
				"APPEND key 1",
				"APPEND key 2",
				"GET key",
				"SET key 1",
				"APPEND key 2",
				"GET key",
			},
			expected: []interface{}{int64(0), int64(1), int64(2), "12", "OK", int64(2), "12"},
		},
	}

	for fuzzType, charset := range types {
		buf := ""
		commands := []string{"DEL key"}
		expected := []interface{}{int64(0)}

		for i := 0; i < 1000; i++ {
			bin := randString(rand.Intn(10)+1, charset)
			buf += bin
			commands = append(commands, "APPEND key "+bin)
			expected = append(expected, int64(len(buf)))
		}

		commands = append(commands, "GET key")
		expected = append(expected, buf)

		testCases = append(testCases, struct {
			name     string
			commands []string
			expected []interface{}
		}{
			name:     "APPEND Fuzzing with " + fuzzType,
			commands: commands,
			expected: expected,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL key")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}
