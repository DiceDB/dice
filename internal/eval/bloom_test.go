package eval

import (
	"bytes"
	"errors"
	"hash"
	"hash/fnv"
	"reflect"
	"testing"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"gotest.tools/v3/assert"
)

func TestBloomFilter(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	// This test only contains some basic checks for all the bloom filter
	// operations like BFINIT, BFADD, BFEXISTS. It assumes that the
	// functions called in the main function are working correctly and
	// are tested independently.
	t.Parallel()

	// BFINIT
	var args []string // empty args
	resp := evalBFINIT(args, store)

	// We're just checking if the resposne is an error or not. This test does not check the type of error. That is kept
	//for different test.
	if bytes.Equal(resp, clientio.RespOK) {
		t.Errorf("BFINIT: invalid response, args: %v - expected an error, got: %s", args, string(resp))
	}

	// BFINIT bf 0.01 10000
	args = append(args, "bf", "0.01", "10000") // Add key, error rate and capacity
	resp = evalBFINIT(args, store)
	if !bytes.Equal(resp, clientio.RespOK) {
		t.Errorf("BFINIT: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespOK), string(resp))
	}

	// BFINIT bf1
	args = []string{"bf1"}
	resp = evalBFINIT(args, store)
	if !bytes.Equal(resp, clientio.RespOK) {
		t.Errorf("BFINIT: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespOK), string(resp))
	}

	// BFADD
	args = []string{"bf"}
	resp = evalBFADD(args, store)
	if bytes.Equal(resp, clientio.RespMinusOne) || bytes.Equal(resp, clientio.RespZero) || bytes.Equal(resp, clientio.RespOne) {
		t.Errorf("BFADD: invalid response, args: %v - expected an error, got: %s", args, string(resp))
	}

	args = []string{"bf", "hello"} // BFADD bf hello
	resp = evalBFADD(args, store)
	if !bytes.Equal(resp, clientio.RespOne) {
		t.Errorf("BFADD: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespOne), string(resp))
	}

	args[1] = "world" // BFADD bf world
	resp = evalBFADD(args, store)
	if !bytes.Equal(resp, clientio.RespOne) {
		t.Errorf("BFADD: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespOne), string(resp))
	}

	args[1] = "hello" // BFADD bf hello
	resp = evalBFADD(args, store)
	if !bytes.Equal(resp, clientio.RespZero) {
		t.Errorf("BFADD: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespZero), string(resp))
	}

	// Try adding element into a non-existing filter
	args = []string{"bf2", "hello"} // BFADD bf2 hello
	resp = evalBFADD(args, store)
	if !bytes.Equal(resp, clientio.RespOne) {
		t.Errorf("BFADD: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespOne), string(resp))
	}

	// BFEXISTS
	args = []string{"bf"}
	resp = evalBFEXISTS(args, store)
	if bytes.Equal(resp, clientio.RespMinusOne) || bytes.Equal(resp, clientio.RespZero) || bytes.Equal(resp, clientio.RespOne) {
		t.Errorf("BFEXISTS: invalid response, args: %v - expected an error, got: %s", args, string(resp))
	}

	args = []string{"bf", "hello"} // BFEXISTS bf hello
	resp = evalBFEXISTS(args, store)
	if !bytes.Equal(resp, clientio.RespOne) {
		t.Errorf("BFEXISTS: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespOne), string(resp))
	}

	args[1] = "hello" // BFEXISTS bf world
	resp = evalBFEXISTS(args, store)
	if !bytes.Equal(resp, clientio.RespOne) {
		t.Errorf("BFEXISTS: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespOne), string(resp))
	}

	args[1] = "programming" // BFEXISTS bf programming
	resp = evalBFEXISTS(args, store)
	if !bytes.Equal(resp, clientio.RespZero) {
		t.Errorf("BFEXISTS: invalid response, args: %v - expected: %s, got error: %s", args, string(clientio.RespZero), string(resp))
	}

	// Try searching for an element in a non-existing filter
	args = []string{"bf3", "hello"} // BFEXISTS bf3 hello
	resp = evalBFEXISTS(args, store)
	if bytes.Equal(resp, clientio.RespMinusOne) || bytes.Equal(resp, clientio.RespZero) || bytes.Equal(resp, clientio.RespOne) {
		t.Errorf("BFEXISTS: invalid response, args: %v - expected an error, got error: %s", args, string(resp))
	}
}

func TestGetOrCreateBloomFilter(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	// Create a key and default opts
	key := "bf"
	opts, _ := newBloomOpts([]string{}, true)

	// Should create a new filter under the key `key`.
	bloom, err := getOrCreateBloomFilter(key, opts, store)
	if bloom == nil || err != nil {
		t.Errorf("nil bloom or non-nil error returned while creating new filter - key: %s, opts: %+v, err: %v", key, opts, err)
	}

	// Should get the filter (which was created above)
	bloom, err = getOrCreateBloomFilter(key, opts, store)
	if bloom == nil || err != nil {
		t.Errorf("nil bloom or non-nil error returned while fetching existing filter - key: %s, opts: %+v, err: %v", key, opts, err)
	}

	// Should get the filter with nil opts
	bloom, err = getOrCreateBloomFilter(key, nil, store)
	if bloom == nil || err != nil {
		t.Errorf("nil bloom or non-nil error returned while fetching existing filter - key: %s, opts: %+v, err: %v", key, opts, err)
	}

	// Should return an error (errInvalidKey) for fetching a bloom filter
	// against a non existing key
	key = "bf1"
	_, err = getOrCreateBloomFilter(key, nil, store)
	if !errors.Is(err, errInvalidKey) {
		t.Errorf("nil or wrong error while fetching non existing bloom filter - key: %s, opts: %+v, err: %v", key, opts, err)
	}
}

func TestUpdateIndexes(t *testing.T) {
	// Create a value, default opts and initialize all params of the filter
	value := "hello"
	opts, _ := newBloomOpts([]string{}, true)
	bloom := newBloomFilter(opts)

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
		{"default values", []string{utils.EmptyStr}, true, &BloomOpts{errorRate: defaultErrorRate, capacity: defaultCapacity}, nil},
		{"should return valid values - 1", []string{"0.01", "1000"}, false, &BloomOpts{errorRate: 0.01, capacity: 1000}, nil},
		{"should return valid values - 2", []string{"0.1", "200"}, false, &BloomOpts{errorRate: 0.1, capacity: 200}, nil},
		{"should return invalid error rate type - 1", []string{"aa", "100"}, false, nil, errInvalidErrorRateType},
		{"should return invalid error rate type - 2", []string{"0.1a", "100"}, false, nil, errInvalidErrorRateType},
		{"should return invalid error rate - 1", []string{"-0.1", "100"}, false, nil, errInvalidErrorRate},
		{"should return invalid error rate - 2", []string{"1.001", "100"}, false, nil, errInvalidErrorRate},
		{"should return invalid capacity type - 1", []string{"0.01", "aa"}, false, nil, errInvalidCapacityType},
		{"should return invalid capacity type - 2", []string{"0.01", "100a"}, false, nil, errInvalidCapacityType},
		{"should return invalid capacity type - 3", []string{"0.01", "-1"}, false, nil, errInvalidCapacityType},
		{"should return invalid capacity - 1", []string{"0.01", "0"}, false, nil, errInvalidCapacity},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opts, err := newBloomOpts(tc.args, tc.useDefaults)
			// Using reflect.DeepEqual as we have pointers to struct and direct value
			// comparision is not possible because of []hash.Hash64 type.
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
	assert.Assert(t, copyBloom != nil, "DeepCopy returned nil, expected a valid copy")

	assert.Assert(t, original.opts.indexes[0] == copyBloom.opts.indexes[0], "Original and copy indexes values should be same")
	assert.Assert(t, original.bitset[0] == copyBloom.bitset[0], "Original and copy bitset values should be same")

	// Verify that changes to the copy do not affect the original
	copyBloom.opts.indexes[0] = 10
	copyBloom.bitset[0] = 0xFF
	assert.Assert(t, original.opts.indexes[0] != copyBloom.opts.indexes[0], "Original and copy indexes should not be linked")
	assert.Assert(t, original.bitset[0] != copyBloom.bitset[0], "Original and copy bitset should not be linked")
}
