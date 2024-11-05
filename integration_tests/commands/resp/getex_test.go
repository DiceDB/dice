package resp

import (
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestGetEx(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	Etime5 := strconv.FormatInt(time.Now().Unix()+5, 10)
	Etime10 := strconv.FormatInt(time.Now().Unix()+10, 10)

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		{
			name:       "GetEx Simple Value",
			commands:   []string{"SET foo bar", "GETEX foo", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", "bar", int64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "GetEx Non-Existent Key",
			commands:   []string{"GETEX foo", "TTL foo"},
			expected:   []interface{}{"(nil)", int64(-2)},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "GetEx with EX option",
			commands:   []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo ex 2", "TTL foo", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), "bar", int64(2), "(nil)", int64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 2 * time.Second, 0},
		},
		{
			name:       "GetEx with PX option",
			commands:   []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo px 2000", "TTL foo", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), "bar", int64(2), "(nil)", int64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 2 * time.Second, 0},
		},
		{
			name:       "GetEx with EX option and invalid value",
			commands:   []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo ex -1", "TTL foo", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), "ERR invalid expire time in 'getex' command", int64(-1), "bar", int64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:       "GetEx with PX option and invalid value",
			commands:   []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo px -1", "TTL foo", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), "ERR invalid expire time in 'getex' command", int64(-1), "bar", int64(-1)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:       "GetEx with EXAT option",
			commands:   []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo exat " + Etime5, "TTL foo", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), "bar", int64(5), "(nil)", int64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 5 * time.Second, 0},
		},
		{
			name:       "GetEx with PXAT option",
			commands:   []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo pxat " + Etime10 + "000", "TTL foo", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), "bar", int64(5), "(nil)", int64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 5 * time.Second, 0},
		},
		{
			name:       "GetEx with EXAT option and invalid value",
			commands:   []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo exat 123123", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), "bar", "(nil)", int64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:       "GetEx with PXAT option and invalid value",
			commands:   []string{"SET foo bar", "GETEX foo", "TTL foo", "GETEX foo pxat 123123", "GETEX foo", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(-1), "bar", "(nil)", int64(-2)},
			assertType: []string{"equal", "equal", "equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:       "GetEx with Persist option",
			commands:   []string{"SET foo bar", "GETEX foo ex 2", "TTL foo", "GETEX foo persist", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(2), "bar", int64(-1)},
			assertType: []string{"equal", "equal", "assert", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name:       "GetEx with multiple expiry options",
			commands:   []string{"SET foo bar", "GETEX foo ex 2 px 123123", "TTL foo", "GETEX foo"},
			expected:   []interface{}{"OK", "ERR syntax error", int64(-1), "bar"},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "GetEx with persist and ex options",
			commands:   []string{"SET foo bar", "GETEX foo ex 2", "TTL foo", "GETEX foo persist ex 2", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(2), "ERR syntax error", int64(2)},
			assertType: []string{"equal", "equal", "assert", "equal", "assert"},
			delay:      []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name:       "GetEx with persist and px options",
			commands:   []string{"SET foo bar", "GETEX foo px 2000", "TTL foo", "GETEX foo px 2000 persist", "TTL foo"},
			expected:   []interface{}{"OK", "bar", int64(2), "ERR syntax error", int64(2)},
			assertType: []string{"equal", "equal", "assert", "equal", "assert"},
			delay:      []time.Duration{0, 0, 0, 0, 0},
		},
		{
			name:       "GetEx with key holding JSON type",
			commands:   []string{"JSON.SET KEY $ \"1\"", "GETEX KEY"},
			expected:   []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:     "GetEx with key holding JSON type with multiple set commands",
			commands: []string{"JSON.SET MJSONKEY $ \"{\"a\":2}\"", "GETEX MJSONKEY", "JSON.SET MJSONKEY $.a \"3\"", "GETEX MJSONKEY"},
			expected: []interface{}{"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
				"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0, 0},
		},
		{
			name:       "GetEx with key holding SET type",
			commands:   []string{"SADD SKEY 1 2 3", "GETEX SKEY"},
			expected:   []interface{}{int64(3), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL foo")

			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := FireCommand(conn, cmd)
				if tc.assertType[i] == "equal" {
					assert.DeepEqual(t, tc.expected[i], result)
				} else if tc.assertType[i] == "assert" {
					assert.Assert(t, result.(int64) <= tc.expected[i].(int64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
