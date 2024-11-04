package clientio_test

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/stretchr/testify/assert"
)

func TestSimpleStringDecode(t *testing.T) {
	cases := map[string]string{
		"+OK\r\n": "OK",
	}
	for k, v := range cases {
		p := clientio.NewRESPParser(bytes.NewBuffer([]byte(k)))
		value, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("error while decoding: %v", k)
		}
		if v != value {
			t.Fail()
		}
	}
}

func TestError(t *testing.T) {
	cases := map[string]string{
		"-Error message\r\n": "Error message",
	}
	for k, v := range cases {
		p := clientio.NewRESPParser(bytes.NewBuffer([]byte(k)))
		value, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("error while decoding: %v", k)
		}
		if v != value {
			t.Fail()
		}
	}
}

func TestInt64(t *testing.T) {
	cases := map[string]int64{
		":0\r\n":    0,
		":1000\r\n": 1000,
	}
	for k, v := range cases {
		p := clientio.NewRESPParser(bytes.NewBuffer([]byte(k)))
		value, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("error while decoding: %v", k)
		}
		if v != value {
			t.Fail()
		}
	}
}

func TestBulkStringDecode(t *testing.T) {
	cases := map[string]string{
		"$5\r\nhello\r\n": "hello",
		"$0\r\n\r\n":      utils.EmptyStr,
	}
	for k, v := range cases {
		p := clientio.NewRESPParser(bytes.NewBuffer([]byte(k)))
		value, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("error while decoding: %v", k)
		}
		if v != value {
			t.Fail()
		}
	}
}

func TestArrayDecode(t *testing.T) {
	cases := map[string][]interface{}{
		"*0\r\n":                                                   {},
		"*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n":                     {"hello", "world"},
		"*3\r\n:1\r\n:2\r\n:3\r\n":                                 {int64(1), int64(2), int64(3)},
		"*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$5\r\nhello\r\n":            {int64(1), int64(2), int64(3), int64(4), "hello"},
		"*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n-World\r\n": {[]int64{int64(1), int64(2), int64(3)}, []interface{}{"Hello", "World"}},
	}
	for k, v := range cases {
		p := clientio.NewRESPParser(bytes.NewBuffer([]byte(k)))
		value, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("error while decoding: %v", v)
		}
		array := value.([]interface{})
		if len(array) != len(v) {
			t.Fail()
		}
		for i := range array {
			if fmt.Sprintf("%v", v[i]) != fmt.Sprintf("%v", array[i]) {
				t.Fail()
			}
		}
	}
}

func TestSimpleStrings(t *testing.T) {
	var b []byte
	var buf = bytes.NewBuffer(b)
	for i := 0; i < 1024; i++ {
		buf.WriteByte('a' + byte(i%26))
		e := clientio.Encode(buf.String(), true)
		p := clientio.NewRESPParser(bytes.NewBuffer(e))
		nv, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("resp parser test failed for value: %v", buf.Bytes())
		}
		if nv != buf.String() {
			t.Errorf("resp parser decoded value mismatch: %v", buf.String())
		}
	}
}

func TestBulkStrings(t *testing.T) {
	var b []byte
	var buf = bytes.NewBuffer(b)
	for i := 0; i < 1024; i++ {
		buf.WriteByte('a' + byte(i%26))
		e := clientio.Encode(buf.String(), false)
		p := clientio.NewRESPParser(bytes.NewBuffer(e))
		nv, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("resp parser test failed for value: %v", buf.Bytes())
		}
		if nv != buf.String() {
			t.Errorf("resp parser decoded value mismatch: %v", buf.String())
		}
	}
}

func TestInt(t *testing.T) {
	for _, v := range []int64{math.MinInt8, math.MinInt16, math.MinInt32, math.MinInt64, 0, math.MaxInt8, math.MaxInt16, math.MaxInt32, math.MaxInt64} {
		e := clientio.Encode(v, false)
		p := clientio.NewRESPParser(bytes.NewBuffer(e))
		nv, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("resp parser test failed for value: %v", v)
		}
		if nv != v {
			t.Errorf("resp parser decoded value mismatch: %v", v)
		}
	}
}

func TestArrayInt(t *testing.T) {
	var b []byte
	var buf = bytes.NewBuffer(b)
	for i := 0; i < 1024; i++ {
		buf.WriteByte('a' + byte(i%26))
		e := clientio.Encode(buf.String(), true)
		p := clientio.NewRESPParser(bytes.NewBuffer(e))
		nv, err := p.DecodeOne()
		if err != nil {
			t.Error(err)
			t.Errorf("resp parser test failed for value: %v", buf.Bytes())
		}
		if nv != buf.String() {
			t.Errorf("resp parser decoded value mismatch: %v", buf.String())
		}
	}
}

func TestBoolean(t *testing.T) {
	tests := []struct {
		input  bool
		output []byte
	}{
		{
			input:  true,
			output: []byte("+true\r\n"),
		},
		{
			input:  false,
			output: []byte("+false\r\n"),
		},
	}

	for _, v := range tests {
		ev := clientio.Encode(v.input, false)
		assert.Equal(t, ev, v.output)
	}
}

func TestInteger(t *testing.T) {
	tests := []struct {
		input  int
		output []byte
	}{
		{
			input:  10,
			output: []byte(":10\r\n"),
		},
		{
			input:  -19,
			output: []byte(":-19\r\n"),
		},
	}

	for _, v := range tests {
		ev := clientio.Encode(v.input, false)
		assert.Equal(t, ev, v.output)
	}
}
