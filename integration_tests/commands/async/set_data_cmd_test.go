package async

import (
	"sort"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func CustomDeepEqual(t *testing.T, a, b interface{}) {
	if a == nil || b == nil {
		assert.DeepEqual(t, a, b)
	}

	switch a.(type) {
	case []any:
		sort.Slice(a.([]any), func(i, j int) bool {
			return a.([]any)[i].(string) < a.([]any)[j].(string)
		})
		sort.Slice(b.([]any), func(i, j int) bool {
			return b.([]any)[i].(string) < b.([]any)[j].(string)
		})
	}

	assert.DeepEqual(t, a, b)
}
func TestSetDataCommand(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name       string
		cmd        []string
		expected   []interface{}
		assertType []string
		delay      []time.Duration
	}{
		// SADD & SDIFF
		{
			name:       "SADD & SDIFF",
			cmd:        []string{"SADD foo bar baz", "SADD foo2 baz bax", "SDIFF foo foo2"},
			expected:   []interface{}{int64(2), int64(2), []any{"bar"}},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "SADD & SDIFF with non existing subsequent key",
			cmd:        []string{"SADD foo bar baz", "SDIFF foo foo2"},
			expected:   []interface{}{int64(2), []any{"bar", "baz"}},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "SADD & SDIFF with wrong key type",
			cmd:        []string{"SET foo bar", "SDIFF foo foo2"},
			expected:   []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "SADD & SDIFF with subsequent key of wrong type",
			cmd:        []string{"SADD foo bar baz", "SET foo2 bar", "SDIFF foo foo2"},
			expected:   []interface{}{int64(2), "OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "SADD & SDIFF with non existing first key",
			cmd:        []string{"SADD foo bar baz", "SDIFF foo2 foo"},
			expected:   []interface{}{int64(2), []any{}},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "SADD & SDIFF with one key",
			cmd:        []string{"SADD foo bar baz", "SDIFF foo"},
			expected:   []interface{}{int64(2), []any{"bar", "baz"}},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		// SADD & SINTER
		{
			name:       "SADD & SINTER",
			cmd:        []string{"SADD foo bar baz", "SADD foo2 baz bax", "SINTER foo foo2"},
			expected:   []interface{}{int64(2), int64(2), []any{"baz"}},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
		{
			name:       "SADD & SINTER with non existing subsequent key",
			cmd:        []string{"SADD foo bar baz", "SINTER foo foo2"},
			expected:   []interface{}{int64(2), []any{}},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "SADD & SINTER with wrong key type",
			cmd:        []string{"SET foo bar", "SINTER foo foo2"},
			expected:   []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal"},
			delay:      []time.Duration{0, 0},
		},
		{
			name:       "SADD & SINTER with subsequent key of wrong type",
			cmd:        []string{"SADD foo bar baz", "SET foo2 bar", "SINTER foo foo2"},
			expected:   []interface{}{int64(2), "OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			assertType: []string{"equal", "equal", "equal"},
			delay:      []time.Duration{0, 0, 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL foo")
			FireCommand(conn, "DEL foo2")
			for i, cmd := range tc.cmd {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := FireCommand(conn, cmd)
				if tc.assertType[i] == "equal" {
					CustomDeepEqual(t, result, tc.expected[i])
				} else if tc.assertType[i] == "assert" {
					assert.Assert(t, result.(int64) <= tc.expected[i].(int64), "Expected %v to be less than or equal to %v", result, tc.expected[i])
				}
			}
		})
	}
}
