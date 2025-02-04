// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"errors"
	"hash"
	"hash/fnv"
	"reflect"
	"testing"

	dstore "github.com/dicedb/dice/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestBloomFilter(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	// This test only contains some basic checks for all the bloom filter
	// operations like BFRESERVE, BFADD, BFEXISTS. It assumes that the
	// functions called in the main function are working correctly and
	// are tested independently.
	t.Parallel()

	// BF.RESERVE
	var args []string // empty args
	resp := evalBFRESERVE(args, store)
	assert.Nil(t, resp.Result)
	assert.Error(t, resp.Error, "ERR wrong number of arguments for 'bf.reserve' command")

	// BF.RESERVE bf 0.01 10000
	args = append(args, "bf", "0.01", "10000") // Add key, error rate and capacity
	resp = evalBFRESERVE(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, OK)

	// BF.RESERVE bf1
	args = []string{"bf1"}
	resp = evalBFRESERVE(args, store)
	assert.Error(t, resp.Error, "ERR wrong number of arguments for 'bf.reserve' command")
	assert.Nil(t, resp.Result)

	// BF.ADD with wrong arguments, must return non-nil error and nil result
	args = []string{"bf"}
	resp = evalBFADD(args, store)
	assert.EqualError(t, resp.Error, "ERR wrong number of arguments for 'bf.add' command")
	assert.Nil(t, resp.Result)

	// BF.ADD bf hello, must return 1 and nil error
	args = []string{"bf", "hello"} // BF.ADD bf hello
	resp = evalBFADD(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, IntegerOne)

	args[1] = "world" // BF.ADD bf world
	resp = evalBFADD(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, IntegerOne)

	args[1] = "hello" // BF.ADD bf hello
	resp = evalBFADD(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, IntegerZero)
	// Try adding element into a non-existing filter
	args = []string{"bf2", "hello"} // BF.ADD bf2 hello
	resp = evalBFADD(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, IntegerOne)

	// BF.EXISTS arity test
	args = []string{"bf"}
	resp = evalBFEXISTS(args, store)
	assert.EqualError(t, resp.Error, "ERR wrong number of arguments for 'bf.exists' command")
	assert.Nil(t, resp.Result)

	args = []string{"bf", "hello"} // BF.EXISTS bf hello
	resp = evalBFEXISTS(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, IntegerOne)

	args[1] = "hello" // BF.EXISTS bf world
	resp = evalBFEXISTS(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, IntegerOne)

	args[1] = "programming" // BF.EXISTS bf programming
	resp = evalBFEXISTS(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, IntegerZero)

	// Try searching for an element in a non-existing filter
	args = []string{"bf3", "hello"} // BF.EXISTS bf3 hello
	resp = evalBFEXISTS(args, store)
	assert.Nil(t, resp.Error)
	assert.ObjectsAreEqualValues(resp.Result, IntegerZero)
}

func TestGetOrCreateBloomFilter(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	// Create a key and default opts
	key := "bf"
	opts := defaultBloomOpts()

	// Should create a new filter under the key `key`.
	bloom, err := GetOrCreateBloomFilter(key, store, opts)
	if bloom == nil || err != nil {
		t.Errorf("nil bloom or non-nil error returned while creating new filter - key: %s, opts: %+v, err: %v", key, opts, err)
	}

	// Should get the filter (which was created above)
	bloom, err = GetOrCreateBloomFilter(key, store, opts)
	if bloom == nil || err != nil {
		t.Errorf("nil bloom or non-nil error returned while fetching existing filter - key: %s, opts: %+v, err: %v", key, opts, err)
	}

	// Should get the filter with nil opts
	bloom, err = GetOrCreateBloomFilter(key, store, nil)
	if bloom == nil || err != nil {
		t.Errorf("nil bloom or non-nil error returned while fetching existing filter - key: %s, opts: %+v, err: %v", key, opts, err)
	}
}

func TestUpdateIndexes(t *testing.T) {
	// Create a value, default opts and initialize all params of the filter
	value := "hello"
	opts := defaultBloomOpts()
	bloom := NewBloomFilter(opts)

	err := opts.updateIndexes(value)
	if err != nil {
		t.Errorf("non-nil error returned from getIndexes - value: %s, opts: %+v", value, opts)
	}

	if len(bloom.opts.indexes) != len(opts.hashFns) {
		t.Errorf("length of indexes does not match with number of hash functions - value: %s, expected: %v, got: %v", value, len(opts.hashFns), len(bloom.opts.indexes))
	}

	for _, index := range bloom.opts.indexes {
		if index >= opts.bits {
			t.Errorf("bit index returned is out of bounds - value: %s, indexes[i]: %d, bound: %d", value, index, opts.bits)
		}
	}
}

func TestBloomOpts(t *testing.T) {
	var testCases = []struct {
		name        string
		args        []string
		useDefaults bool
		response    *BloomOpts
		err         error
	}{
		{"should return valid values - 1", []string{"0.01", "1000"}, false, &BloomOpts{errorRate: 0.01, capacity: 1000}, nil},
		{"should return valid values - 2", []string{"0.1", "200"}, false, &BloomOpts{errorRate: 0.1, capacity: 200}, nil},
		{"should return invalid error rate type - 1", []string{"aa", "100"}, false, nil, errInvalidErrorRate},
		{"should return invalid error rate type - 2", []string{"0.1a", "100"}, false, nil, errInvalidErrorRate},
		{"should return invalid error rate - 1", []string{"-0.1", "100"}, false, nil, errInvalidRangeErrorRateType},
		{"should return invalid error rate - 2", []string{"1.001", "100"}, false, nil, errInvalidRangeErrorRateType},
		{"should return invalid capacity type - 1", []string{"0.01", "aa"}, false, nil, errInvalidCapacityType},
		{"should return invalid capacity type - 2", []string{"0.01", "100a"}, false, nil, errInvalidCapacityType},
		{"should return invalid capacity type - 3", []string{"0.01", "-1"}, false, nil, errNonPositiveCapacity},
		{"should return invalid capacity - 1", []string{"0.01", "0"}, false, nil, errNonPositiveCapacity},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opts, err := newBloomOpts(tc.args)
			// Using reflect.DeepEqual as we have pointers to struct and direct value
			// comparison is not possible because of []hash.Hash64 type.
			if !reflect.DeepEqual(opts, tc.response) {
				t.Errorf("invalid response in %s - expected: %v, got: %v", t.Name(), tc.response, opts)
			}

			if !errors.Is(err, tc.err) {
				t.Errorf("invalid error in %s - expected: %v, got: %v", t.Name(), tc.err, err)
			}
		})
	}
}

func TestIsBitSet(t *testing.T) {
	buf := []byte{170, 43} // 10101010 00101011
	var testCases = []struct {
		name     string
		index    uint64
		expected bool
	}{
		{"Handle index equal to length", 16, false},
		{"Handle index more than length", 17, false},
		{"Handle start bit 1", 0, true},
		{"Handle start bit 2", 8, false},
		{"Handle mid bit 1", 3, false},
		{"Handle mid bit 2", 4, true},
		{"Handle mid bit 3", 11, false},
		{"Handle mid bit 4", 14, true},
		{"Handle end bit 1", 7, false},
		{"Handle end bit 2", 15, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := isBitSet(buf, tc.index)
			if actual != tc.expected {
				t.Errorf("error in %s - expected: %t, got: %t", t.Name(), tc.expected, actual)
			}
		})
	}
}

func TestSetBit(t *testing.T) {
	buf := []byte{170, 43} // 10101010 00101011
	var testCases = []struct {
		name     string
		index    uint64
		expected bool
	}{
		{"Handle index equal to length", 16, false},
		{"Handle index more than length", 17, false},
		{"Handle start bit 1", 0, true}, // 10101010 00101011
		{"Handle start bit 2", 8, true}, // 10101010 10101011
		{"Handle mid bit 1", 3, true},   // 10111010 10101011
		{"Handle mid bit 2", 4, true},   // 10111010 10101011
		{"Handle mid bit 3", 11, true},  // 10111010 10111011
		{"Handle mid bit 4", 14, true},  // 10111010 10111011
		{"Handle end bit 1", 7, true},   // 10111011 10111011
		{"Handle end bit 2", 15, true},  // 10111011 10111011
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Set the bit first and then try to read it
			setBit(buf, tc.index)
			actual := isBitSet(buf, tc.index)
			if actual != tc.expected {
				t.Errorf("error in %s - expected: %t, got: %t", t.Name(), tc.expected, actual)
			}
		})
	}

	// The final values are 10111011 (=187)
	expected1, expected2 := 187, 187
	if int(buf[0]) != expected1 || int(buf[1]) != expected2 {
		t.Errorf("error in %s while comparing final buffer values - expected: [%d, %d], got: [%d, %d]", t.Name(), expected1, expected2, int(buf[0]), int(buf[1]))
	}
}

func TestBloomDeepCopy(t *testing.T) {
	// mock data
	originalOpts := &BloomOpts{
		errorRate: 0.01,
		capacity:  1000,
		bits:      8000,
		bpe:       8.0,
		hashFns: []hash.Hash64{
			fnv.New64a(),
			fnv.New64(),
		},
		indexes: []uint64{1, 2, 3, 4, 5},
	}

	original := &Bloom{
		opts:   originalOpts,
		bitset: []byte{0x0F, 0xF0, 0xAA, 0x55},
	}

	// Create a deep copy of the Bloom filter
	copyBloom := original.DeepCopy()

	// Verify that the copy is not nil
	assert.NotNil(t, copyBloom, "DeepCopy returned nil, expected a valid copy")

	assert.True(t, original.opts.indexes[0] == copyBloom.opts.indexes[0], "Original and copy indexes values should be same")
	assert.True(t, original.bitset[0] == copyBloom.bitset[0], "Original and copy bitset values should be same")

	// Verify that changes to the copy do not affect the original
	copyBloom.opts.indexes[0] = 10
	copyBloom.bitset[0] = 0xFF
	assert.True(t, original.opts.indexes[0] != copyBloom.opts.indexes[0], "Original and copy indexes should not be linked")
	assert.True(t, original.bitset[0] != copyBloom.bitset[0], "Original and copy bitset should not be linked")
}
