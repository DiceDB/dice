package tests

import (
	"testing"

	"github.com/dicedb/dice/core/xencoding"
)

func TestXencodingInt(t *testing.T) {
	var tests []uint64

	tests = append(tests, 0)
	for i := 0; i < 64; i++ {
		tests = append(tests, 1<<i)
		tests = append(tests, (1<<i)-1)
		tests = append(tests, (1<<i)+1)
	}

	for _, v := range tests {
		if v != xencoding.XDecodeUInt(xencoding.XEncodeUInt(v)) {
			t.Errorf("xencoding failed for value: %d\n", v)
		}
	}

	for k, v := range map[uint64]int{
		0:   1,
		127: 1,
		128: 2,
		129: 2,
	} {
		b := xencoding.XEncodeUInt(k)
		if len(b) != v {
			t.Errorf("xencoding for integer value %d failed. encoded length should be: %d, but found %d\n", k, v, len(b))
		}
	}
}
