package async

import (
	"testing"
	"time"

	testifyAssert "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"
)

func TestBitfield(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	defer FireCommand(conn, "FLUSHDB") // clean up after all test cases

	testCases := []struct {
		Name     string
		Setup    []string
		Commands []string
		Expected []interface{}
		Delay    []time.Duration
		CleanUp  []string
	}{
		{
			Name:     "BITFIELD signed SET and GET basics",
			Setup:    []string{},
			Commands: []string{"bitfield bits set i8 0 -100", "bitfield bits set i8 0 101", "bitfield bits get i8 0"},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(-100)}, []interface{}{int64(101)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD unsigned SET and GET basics",
			Setup:    []string{},
			Commands: []string{"bitfield bits set u8 0 255", "bitfield bits set u8 0 100", "bitfield bits get u8 0"},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(255)}, []interface{}{int64(100)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD signed SET and GET together",
			Setup:    []string{},
			Commands: []string{"bitfield bits set i8 0 255 set i8 0 100 get i8 0"},
			Expected: []interface{}{[]interface{}{int64(0), int64(-1), int64(100)}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD unsigned with SET, GET and INCRBY arguments",
			Setup:    []string{},
			Commands: []string{"bitfield bits set u8 0 255 incrby u8 0 100 get u8 0"},
			Expected: []interface{}{[]interface{}{int64(0), int64(99), int64(99)}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD with only key as argument",
			Setup:    []string{},
			Commands: []string{"bitfield bits"},
			Expected: []interface{}{[]interface{}{}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:  "BITFIELD #<idx> form",
			Setup: []string{},
			Commands: []string{
				"bitfield bits set u8 #0 65",
				"bitfield bits set u8 #1 66",
				"bitfield bits set u8 #2 67",
				"get bits",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(0)}, []interface{}{int64(0)}, "ABC"},
			Delay:    []time.Duration{0, 0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:  "BITFIELD basic INCRBY form",
			Setup: []string{},
			Commands: []string{
				"bitfield bits set u8 #0 10",
				"bitfield bits incrby u8 #0 100",
				"bitfield bits incrby u8 #0 100",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(110)}, []interface{}{int64(210)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:  "BITFIELD chaining of multiple commands",
			Setup: []string{},
			Commands: []string{
				"bitfield bits set u8 #0 10",
				"bitfield bits incrby u8 #0 100 incrby u8 #0 100",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(110), int64(210)}},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:  "BITFIELD unsigned overflow wrap",
			Setup: []string{},
			Commands: []string{
				"bitfield bits set u8 #0 100",
				"bitfield bits overflow wrap incrby u8 #0 257",
				"bitfield bits get u8 #0",
				"bitfield bits overflow wrap incrby u8 #0 255",
				"bitfield bits get u8 #0",
			},
			Expected: []interface{}{
				[]interface{}{int64(0)},
				[]interface{}{int64(101)},
				[]interface{}{int64(101)},
				[]interface{}{int64(100)},
				[]interface{}{int64(100)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"DEL bits"},
		},
		{
			Name:  "BITFIELD unsigned overflow sat",
			Setup: []string{},
			Commands: []string{
				"bitfield bits set u8 #0 100",
				"bitfield bits overflow sat incrby u8 #0 257",
				"bitfield bits get u8 #0",
				"bitfield bits overflow sat incrby u8 #0 -255",
				"bitfield bits get u8 #0",
			},
			Expected: []interface{}{
				[]interface{}{int64(0)},
				[]interface{}{int64(255)},
				[]interface{}{int64(255)},
				[]interface{}{int64(0)},
				[]interface{}{int64(0)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"DEL bits"},
		},
		{
			Name:  "BITFIELD signed overflow wrap",
			Setup: []string{},
			Commands: []string{
				"bitfield bits set i8 #0 100",
				"bitfield bits overflow wrap incrby i8 #0 257",
				"bitfield bits get i8 #0",
				"bitfield bits overflow wrap incrby i8 #0 255",
				"bitfield bits get i8 #0",
			},
			Expected: []interface{}{
				[]interface{}{int64(0)},
				[]interface{}{int64(101)},
				[]interface{}{int64(101)},
				[]interface{}{int64(100)},
				[]interface{}{int64(100)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"DEL bits"},
		},
		{
			Name:  "BITFIELD signed overflow sat",
			Setup: []string{},
			Commands: []string{
				"bitfield bits set u8 #0 100",
				"bitfield bits overflow sat incrby i8 #0 257",
				"bitfield bits get i8 #0",
				"bitfield bits overflow sat incrby i8 #0 -255",
				"bitfield bits get i8 #0",
			},
			Expected: []interface{}{
				[]interface{}{int64(0)},
				[]interface{}{int64(127)},
				[]interface{}{int64(127)},
				[]interface{}{int64(-128)},
				[]interface{}{int64(-128)},
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD regression 1",
			Setup:    []string{},
			Commands: []string{"set bits 1", "bitfield bits get u1 0"},
			Expected: []interface{}{"OK", []interface{}{int64(0)}},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:  "BITFIELD regression 2",
			Setup: []string{},
			Commands: []string{
				"bitfield mystring set i8 0 10",
				"bitfield mystring set i8 64 10",
				"bitfield mystring incrby i8 10 99900",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(0)}, []interface{}{int64(60)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL mystring"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			for _, cmd := range tc.Setup {
				assert.Equal(t, FireCommand(conn, cmd), "OK")
			}

			for i := 0; i < len(tc.Commands); i++ {
				if tc.Delay[i] > 0 {
					time.Sleep(tc.Delay[i])
				}
				result := FireCommand(conn, tc.Commands[i])
				expected := tc.Expected[i]
				testifyAssert.Equal(t, expected, result)
			}

			for _, cmd := range tc.CleanUp {
				FireCommand(conn, cmd)
			}
		})
	}
}
