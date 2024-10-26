package async

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestObjectCommand(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	defer FireCommand(conn, "FLUSHDB")
	simpleJSON := `{"name":"John","age":30}`

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
		cleanup    []string
	}{
		{
			name:       "Object Idletime",
			commands:   []string{"SET foo bar", "OBJECT IDLETIME foo", "OBJECT IDLETIME foo", "TOUCH foo", "OBJECT IDLETIME foo"},
			expected:   []interface{}{"OK", int64(2), int64(3), int64(1), int64(0)},
			assertType: []string{"equal", "assert", "assert", "equal", "assert"},
			delay:      []time.Duration{0, 2 * time.Second, 3 * time.Second, 0, 0},
			cleanup:    []string{"DEL foo"},
		},
		{
			name:       "Object Encoding check for raw",
			commands:   []string{"SET foo foobarfoobarfoobarfoobarfoobarfoobarfoobarfoobarfoobarfoobarfoobarfoobarfoobarfoobar", "OBJECT ENCODING foo"},
			expected:   []interface{}{"OK", "raw"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
			cleanup:    []string{"DEL foo"},
		},
		{
			name:       "Object Encoding check for int",
			commands:   []string{"SET foo 1", "OBJECT ENCODING foo"},
			expected:   []interface{}{"OK", "int"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
			cleanup:    []string{"DEL foo"},
		},
		{
			name:       "Object Encoding check for embstr",
			commands:   []string{"SET foo bar", "OBJECT ENCODING foo"},
			expected:   []interface{}{"OK", "embstr"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
			cleanup:    []string{"DEL foo"},
		},
		{
			name:       "Object Encoding check for deque",
			commands:   []string{"LPUSH listKey 'value1'", "LPUSH listKey 'value2'", "OBJECT ENCODING listKey"},
			expected:   []interface{}{int64(1), int64(2), "deque"},
			assertType: []string{"assert", "assert", "equal"},
			delay:      []time.Duration{0, 0, 0},
			cleanup:    []string{"DEL listKey"},
		},
		{
			name:       "Object Encoding check for bf",
			commands:   []string{"BF.ADD bloomkey value1", "BF.ADD bloomkey value2", "OBJECT ENCODING bloomkey"},
			expected:   []interface{}{int64(1), int64(1), "bf"},
			assertType: []string{"assert", "assert", "equal"},
			delay:      []time.Duration{0, 0, 0},
			cleanup:    []string{"DEL bloomkey"},
		},
		{
			name:       "Object Encoding check for json",
			commands:   []string{`JSON.SET k1 $ ` + simpleJSON, "OBJECT ENCODING k1"},
			expected:   []interface{}{"OK", "json"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
			cleanup:    []string{"DEL k1"},
		},
		{
			name:       "Object Encoding check for bytearray",
			commands:   []string{"SETBIT kbitset 0 1", "SETBIT kbitset 1 0", "SETBIT kbitset 2 1", "OBJECT ENCODING kbitset"},
			expected:   []interface{}{int64(0), int64(0), int64(0), "bytearray"},
			assertType: []string{"assert", "assert", "assert", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
			cleanup:    []string{"DEL kbitset"},
		},
		{
			name:       "Object Encoding check for hashmap",
			commands:   []string{"HSET hashKey hKey hValue", "OBJECT ENCODING hashKey"},
			expected:   []interface{}{int64(1), "hashmap"},
			assertType: []string{"assert", "equal"},
			delay:      []time.Duration{0, 0},
			cleanup:    []string{"DEL hashKey"},
		},
		{
			name:       "Object Encoding check for btree",
			commands:   []string{"ZADD btreekey 1 'member1' 2 'member2'", "OBJECT ENCODING btreekey"},
			expected:   []interface{}{int64(2), "btree"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
			cleanup:    []string{"DEL btreekey"},
		},
		{
			name:       "Object Encoding check for setstr",
			commands:   []string{"SADD skey one two three", "OBJECT ENCODING skey"},
			expected:   []interface{}{int64(3), "setstr"},
			assertType: []string{"assert", "equal"},
			delay:      []time.Duration{0, 0},
			cleanup:    []string{"DEL skey"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"foo"}, store)
			FireCommand(conn, "DEL foo")

			for i, cmd := range tc.commands {
				if tc.delay[i] != 0 {
					time.Sleep(tc.delay[i])
				}

				result := FireCommand(conn, cmd)

				fmt.Println(cmd, result, tc.expected[i])
				if tc.assertType[i] == "equal" {
					assert.DeepEqual(t, tc.expected[i], result)
				} else {
					assert.Assert(t, result.(int64) >= tc.expected[i].(int64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
			for _, cmd := range tc.cleanup { // run cleanup
				FireCommand(conn, cmd)
			}
		})
	}
}
