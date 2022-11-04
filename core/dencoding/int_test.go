package dencoding_test

import (
	"testing"

	"github.com/dicedb/dice/core/dencoding"
	"github.com/dicedb/dice/testutils"
)

func TestDencodingInt(t *testing.T) {
	var tests []uint64

	tests = append(tests, 0)
	for i := 0; i < 64; i++ {
		tests = append(tests, 1<<i)
		tests = append(tests, (1<<i)-1)
		tests = append(tests, (1<<i)+1)
	}

	for _, v := range tests {
		if v != dencoding.DecodeUInt(dencoding.EncodeUInt(v)) {
			t.Errorf("dencoding failed for value: %d\n", v)
		}
	}

	for k, v := range map[uint64]int{
		0:   1,
		127: 1,
		128: 2,
		129: 2,
	} {
		b := dencoding.EncodeUInt(k)
		if len(b) != v {
			t.Errorf("dencoding for integer value %d failed. encoded length should be: %d, but found %d\n", k, v, len(b))
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
		obs := dencoding.EncodeUInt(uint64(k))
		if !testutils.EqualByteSlice(obs, v) {
			t.Errorf("dencoding for integer value %d failed. should be: %d, but found %d\n", k, v, obs)
		}
	}
}
