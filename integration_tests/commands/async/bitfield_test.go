package async

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBitfield(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "FLUSHDB")
	defer FireCommand(conn, "FLUSHDB") // clean up after all test cases
	syntaxErrMsg := "ERR syntax error"
	bitFieldTypeErrMsg := "ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is."
	integerErrMsg := "ERR value is not an integer or out of range"
	overflowErrMsg := "ERR Invalid OVERFLOW type specified"

	testCases := []struct {
		Name     string
		Commands []string
		Expected []interface{}
		Delay    []time.Duration
		CleanUp  []string
	}{
		{
			Name:     "BITFIELD Arity Check",
			Commands: []string{"bitfield"},
			Expected: []interface{}{"ERR wrong number of arguments for 'bitfield' command"},
			Delay:    []time.Duration{0},
			CleanUp:  []string{},
		},
		{
			Name:     "BITFIELD on unsupported type of SET",
			Commands: []string{"SADD bits a b c", "bitfield bits"},
			Expected: []interface{}{int64(3), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD on unsupported type of JSON",
			Commands: []string{"json.set bits $ 1", "bitfield bits"},
			Expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD on unsupported type of HSET",
			Commands: []string{"HSET bits a 1", "bitfield bits"},
			Expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD with syntax errors",
			Commands: []string{
				"bitfield bits set u8 0 255 incrby u8 0 100 get u8",
				"bitfield bits set a8 0 255 incrby u8 0 100 get u8",
				"bitfield bits set u8 a 255 incrby u8 0 100 get u8",
				"bitfield bits set u8 0 255 incrby u8 0 100 overflow wraap",
				"bitfield bits set u8 0 incrby u8 0 100 get u8 288",
			},
			Expected: []interface{}{
				syntaxErrMsg,
				bitFieldTypeErrMsg,
				"ERR bit offset is not an integer or out of range",
				overflowErrMsg,
				integerErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0, 0},
			CleanUp: []string{"Del bits"},
		},
		{
			Name:     "BITFIELD signed SET and GET basics",
			Commands: []string{"bitfield bits set i8 0 -100", "bitfield bits set i8 0 101", "bitfield bits get i8 0"},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(-100)}, []interface{}{int64(101)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD unsigned SET and GET basics",
			Commands: []string{"bitfield bits set u8 0 255", "bitfield bits set u8 0 100", "bitfield bits get u8 0"},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(255)}, []interface{}{int64(100)}},
			Delay:    []time.Duration{0, 0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD signed SET and GET together",
			Commands: []string{"bitfield bits set i8 0 255 set i8 0 100 get i8 0"},
			Expected: []interface{}{[]interface{}{int64(0), int64(-1), int64(100)}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD unsigned with SET, GET and INCRBY arguments",
			Commands: []string{"bitfield bits set u8 0 255 incrby u8 0 100 get u8 0"},
			Expected: []interface{}{[]interface{}{int64(0), int64(99), int64(99)}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD with only key as argument",
			Commands: []string{"bitfield bits"},
			Expected: []interface{}{[]interface{}{}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD #<idx> form",
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
			Name: "BITFIELD basic INCRBY form",
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
			Name: "BITFIELD chaining of multiple commands",
			Commands: []string{
				"bitfield bits set u8 #0 10",
				"bitfield bits incrby u8 #0 100 incrby u8 #0 100",
			},
			Expected: []interface{}{[]interface{}{int64(0)}, []interface{}{int64(110), int64(210)}},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD unsigned overflow wrap",
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
			Name: "BITFIELD unsigned overflow sat",
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
			Name: "BITFIELD signed overflow wrap",
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
			Name: "BITFIELD signed overflow sat",
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
			Commands: []string{"set bits 1", "bitfield bits get u1 0"},
			Expected: []interface{}{"OK", []interface{}{int64(0)}},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD regression 2",
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

			for i := 0; i < len(tc.Commands); i++ {
				if tc.Delay[i] > 0 {
					time.Sleep(tc.Delay[i])
				}
				result := FireCommand(conn, tc.Commands[i])
				expected := tc.Expected[i]
				assert.Equal(t, expected, result)
			}

			for _, cmd := range tc.CleanUp {
				FireCommand(conn, cmd)
			}
		})
	}
}

func TestBitfieldRO(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	FireCommand(conn, "FLUSHDB")
	defer FireCommand(conn, "FLUSHDB")

	syntaxErrMsg := "ERR syntax error"
	bitFieldTypeErrMsg := "ERR Invalid bitfield type. Use something like i16 u8. Note that u64 is not supported but i64 is."
	unsupportedCmdErrMsg := "ERR BITFIELD_RO only supports the GET subcommand"

	testCases := []struct {
		Name     string
		Commands []string
		Expected []interface{}
		Delay    []time.Duration
		CleanUp  []string
	}{
		{
			Name:     "BITFIELD_RO Arity Check",
			Commands: []string{"bitfield_ro"},
			Expected: []interface{}{"ERR wrong number of arguments for 'bitfield_ro' command"},
			Delay:    []time.Duration{0},
			CleanUp:  []string{},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of SET",
			Commands: []string{"SADD bits a b c", "bitfield_ro bits"},
			Expected: []interface{}{int64(3), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of JSON",
			Commands: []string{"json.set bits $ 1", "bitfield_ro bits"},
			Expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name:     "BITFIELD_RO on unsupported type of HSET",
			Commands: []string{"HSET bits a 1", "bitfield_ro bits"},
			Expected: []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			Delay:    []time.Duration{0, 0},
			CleanUp:  []string{"DEL bits"},
		},
		{
			Name: "BITFIELD_RO with unsupported commands",
			Commands: []string{
				"bitfield_ro bits set u8 0 255",
				"bitfield_ro bits incrby u8 0 100",
			},
			Expected: []interface{}{
				unsupportedCmdErrMsg,
				unsupportedCmdErrMsg,
			},
			Delay:   []time.Duration{0, 0},
			CleanUp: []string{"Del bits"},
		},
		{
			Name: "BITFIELD_RO with syntax error",
			Commands: []string{
				"set bits 1",
				"bitfield_ro bits get u8",
				"bitfield_ro bits get",
				"bitfield_ro bits get somethingrandom",
			},
			Expected: []interface{}{
				"OK",
				syntaxErrMsg,
				syntaxErrMsg,
				syntaxErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0},
			CleanUp: []string{"Del bits"},
		},
		{
			Name: "BITFIELD_RO with invalid bitfield type",
			Commands: []string{
				"set bits 1",
				"bitfield_ro bits get a8 0",
				"bitfield_ro bits get s8 0",
				"bitfield_ro bits get somethingrandom 0",
			},
			Expected: []interface{}{
				"OK",
				bitFieldTypeErrMsg,
				bitFieldTypeErrMsg,
				bitFieldTypeErrMsg,
			},
			Delay:   []time.Duration{0, 0, 0, 0},
			CleanUp: []string{"Del bits"},
		},
		{
			Name:     "BITFIELD_RO with only key as argument",
			Commands: []string{"bitfield_ro bits"},
			Expected: []interface{}{[]interface{}{}},
			Delay:    []time.Duration{0},
			CleanUp:  []string{"DEL bits"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			for i := 0; i < len(tc.Commands); i++ {
				if tc.Delay[i] > 0 {
					time.Sleep(tc.Delay[i])
				}
				result := FireCommand(conn, tc.Commands[i])
				expected := tc.Expected[i]
				assert.Equal(t, expected, result)
			}

			for _, cmd := range tc.CleanUp {
				FireCommand(conn, cmd)
			}
		})
	}
}
