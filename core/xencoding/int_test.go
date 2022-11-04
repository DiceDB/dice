package xencoding_test

import (
	"testing"

	"github.com/dicedb/dice/core/xencoding"
	"github.com/dicedb/dice/testutils"
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

	for k, v := range map[int][]byte{
		1:   {0b00000001},
		2:   {0b00000010},
		127: {0b01111111},
		128: {0b10000000, 0b00000001},
		129: {0b10000001, 0b00000001},
		130: {0b10000010, 0b00000001},
		131: {0b10000011, 0b00000001},
	} {
		obs := xencoding.XEncodeUInt(uint64(k))
		if !testutils.EqualByteSlice(obs, v) {
			t.Errorf("xencoding for integer value %d failed. should be: %d, but found %d\n", k, v, obs)
		}
	}
}
