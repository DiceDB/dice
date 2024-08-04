package dencoding_test

import (
	"math"
	"testing"

	"github.com/dicedb/dice/core/dencoding"
	"github.com/dicedb/dice/testutils"
)

func TestDencodingUInt(t *testing.T) {
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

func TestDencodingInt(t *testing.T) {
	var tests []int64

	tests = append(tests, 0)
	for i := 0; i < 64; i++ {
		tests = append(tests, 1<<i)
		tests = append(tests, (1<<i)-1)
		tests = append(tests, (1<<i)+1)
		tests = append(tests, -1<<i)
		tests = append(tests, -(1<<i)-1)
		tests = append(tests, -(1<<i)+1)
	}

	for _, v := range tests {
		if v != dencoding.DecodeInt(dencoding.EncodeInt(v)) {
			t.Errorf("dencoding failed for value: %d\n", v)
		}
	}

	for k, v := range map[int64]int{
		0:   1,
		127: 2,
		128: 2,
		129: 2,
	} {
		b := dencoding.EncodeInt(int64(k))
		if len(b) != v {
			t.Errorf("dencoding for integer value %d failed. encoded length should be: %d, but found %d\n", k, v, len(b))
		}
	}

	for k, v := range map[int][]byte{
		0:    {0x00},
		1:    {0x02},
		-1:   {0x01},
		63:   {0x7E},
		-64:  {0x7F},
		64:   {0x80, 0x01},
		-65:  {0x81, 0x01},
		127:  {0xFE, 0x01},
		128:  {0x80, 0x02},
		-128: {0xFF, 0x01},
		-129: {0x81, 0x02},
	} {
		obs := dencoding.EncodeInt(int64(k))
		if !testutils.EqualByteSlice(obs, v) {
			t.Errorf("dencoding for integer value %d failed. should be: %d, but found %d\n", k, v, obs)
		}
	}
}

func TestEncodeDecodeInt64Min(t *testing.T) {
	value := int64(math.MinInt64)
	encoded := dencoding.EncodeInt(value)
	decoded := dencoding.DecodeInt(encoded)
	if decoded != value {
		t.Errorf("DecodeInt(%v) = %d; want %d", encoded, decoded, value)
	}
}

func BenchmarkEncodeDecodeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		value := int64(i % math.MaxInt64)
		encoded := dencoding.EncodeInt(value)
		decoded := dencoding.DecodeInt(encoded)
		if decoded != value {
			b.Errorf("DecodeInt(%v) = %d; want %d", encoded, decoded, value)
		}
	}
}
