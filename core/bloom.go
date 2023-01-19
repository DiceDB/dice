package core

import (
	"errors"
	"fmt"
	"hash"
	"math"
	"math/rand"
	"strconv"

	"github.com/twmb/murmur3"
)

const (
	defaultErrorRate float64 = 0.01
	defaultCapacity  uint64  = 1024
)

var (
	ln2      float64 = math.Log(2)
	ln2Power float64 = ln2 * ln2
)

var (
	errWrongArgs            = errors.New("ERR wrong number of arguments")
	errInvalidErrorRateType = errors.New("ERR only float values can be provided for error rate")
	errInvalidErrorRate     = errors.New("ERR invalid error rate value provided")
	errInvalidCapacityType  = errors.New("ERR only integer values can be provided for capacity")
	errInvalidCapacity      = errors.New("ERR invalid capacity value provided")

	errInvalidKey = errors.New("ERR invalid key: no bloom filter found")

	errEmptyValue   = errors.New("ERR empty value provided")
	errUnableToHash = errors.New("ERR unable to hash given value")
)

type BloomOpts struct {
	errorRate float64 // desired error rate (the false positive rate) of the filter
	capacity  uint64  // number of expected entries to be added to the filter

	bits    uint64        // total number of bits reserved for the filter
	hashFns []hash.Hash64 // array of hash functions
	bpe     float64       // bits per element

	// indexes slice will hold the indexes, representing bits to be set/read and
	// is under the assumption that it's consumed at only 1 place at a time. Add
	// a lock when multiple clients can be supported.
	indexes []uint64
}

type Bloom struct {
	opts   *BloomOpts // options for the bloom filter
	bitset []byte     // underlying bit representation
}

// newBloomOpts extracts the user defined values from `args`. It falls back to
// default values if `useDefaults` is set to true. Using those values, it
// creates and returns the options for bloom filter.
func newBloomOpts(args []string, useDefaults bool) (*BloomOpts, error) {
	if useDefaults {
		return &BloomOpts{errorRate: defaultErrorRate, capacity: defaultCapacity}, nil
	}

	errorRate, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return nil, errInvalidErrorRateType
	}

	if errorRate <= 0 || errorRate >= 1.0 {
		return nil, errInvalidErrorRate
	}

	capacity, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return nil, errInvalidCapacityType
	}

	if capacity < 1 {
		return nil, errInvalidCapacity
	}

	return &BloomOpts{errorRate: errorRate, capacity: capacity}, nil
}

// newBloomFilter creates and returns a new filter. It is responsible for initializing the
// underlying bit array.
func newBloomFilter(opts *BloomOpts) *Bloom {
	// Calculate bits per element
	// 		bpe = -log(errorRate)/ln(2)^2
	num := -1 * math.Log(opts.errorRate)
	opts.bpe = num / ln2Power

	// Calculate the number of hash functions to be used
	// 		k = ceil(ln(2) * bpe)
	k := math.Ceil(ln2 * opts.bpe)
	opts.hashFns = make([]hash.Hash64, int(k))

	// Initialize hash functions with random seeds
	for i := 0; i < int(k); i++ {
		opts.hashFns[i] = murmur3.SeedNew64(rand.Uint64())
	}

	// initialize the common slice for storing indexes of bits to be set
	opts.indexes = make([]uint64, len(opts.hashFns))

	// Calculate the number of bytes to be used
	// 		bits = k * entries / ln(2)
	//		bytes = bits * 8
	bits := uint64(math.Ceil((k * float64(opts.capacity)) / ln2))
	var bytes uint64
	if bits%8 == 0 {
		bytes = bits / 8
	} else {
		bytes = (bits / 8) + 1
	}
	opts.bits = bytes * 8

	bitset := make([]byte, bytes)

	return &Bloom{opts, bitset}
}

func (b *Bloom) info(name string) string {
	info := ""
	if name != "" {
		info = "name: " + name + ", "
	}
	info += fmt.Sprintf("error rate: %f, ", b.opts.errorRate)
	info += fmt.Sprintf("capacity: %d, ", b.opts.capacity)
	info += fmt.Sprintf("total bits reserved: %d, ", b.opts.bits)
	info += fmt.Sprintf("bits per element: %f, ", b.opts.bpe)
	info += fmt.Sprintf("hash functions: %d", len(b.opts.hashFns))

	return info
}

// add adds a new entry for `value` in the filter. It hashes the given
// value and sets the bit of the underlying bitset. Returns "-1" in
// case of errors, "0" if all the bits were already set and "1" if
// atleast 1 new bit was set.
func (b *Bloom) add(value string) ([]byte, error) {
	// We're sure that empty values will be handled upper functions itself.
	// This is just a property check for the bloom struct.
	if value == "" {
		return RESP_MINUS_1, errEmptyValue
	}

	// Update the indexes where bits are supposed to be set
	err := b.opts.updateIndexes(value)
	if err != nil {
		fmt.Println("error in getting indexes for value:", value, "err:", err)
		return RESP_MINUS_1, errUnableToHash
	}

	// Set the bits and keep a count of already set ones
	count := 0
	for _, v := range b.opts.indexes {
		if isBitSet(b.bitset, v) {
			count++
		} else {
			setBit(b.bitset, v)
		}
	}

	if count == len(b.opts.indexes) {
		// All the bits were already set, return 0 in that case.
		return RESP_ZERO, nil
	}

	return RESP_ONE, nil
}

// exists checks if the given `value` exists in the filter or not.
// It hashes the given value and checks if the bits are set or not in
// the underlying bitset. Returns "-1" in case of errors, "0" if the
// element surely does not exist in the filter, and "1" if the element
// may or may not exist in the filter.
func (b *Bloom) exists(value string) ([]byte, error) {
	// We're sure that empty values will be handled upper functions itself.
	// This is just a property check for the bloom struct.
	if value == "" {
		return RESP_MINUS_1, errEmptyValue
	}

	// Update the indexes where bits are supposed to be set
	err := b.opts.updateIndexes(value)
	if err != nil {
		fmt.Println("error in getting indexes for value:", value, "err:", err)
		return RESP_MINUS_1, errUnableToHash
	}

	// Check if all the bits at given indexes are set or not
	// Ideally if the element is present, we should find all set bits.
	for _, v := range b.opts.indexes {
		if !isBitSet(b.bitset, v) {
			// Return with "0" as we found one non-set bit (which is enough to conclude)
			return RESP_ZERO, nil
		}
	}

	// We reached here, which means the element may exist in the filter. Return "1" now.
	return RESP_ONE, nil
}

// updateIndexes updates the list with indexes where bits are supposed to be
// set (to 1) or read in/from the underlying array. It uses the set hash function
// against the given `value` and caps the index with the total number of bits.
func (opts *BloomOpts) updateIndexes(value string) error {
	// Iterate through the hash functions and get indexes
	for i := 0; i < len(opts.hashFns); i++ {
		fn := opts.hashFns[i]
		fn.Reset()

		if _, err := fn.Write([]byte(value)); err != nil {
			return err
		}

		// Save the index capped by total number of bits in the underlying array
		opts.indexes[i] = fn.Sum64() % opts.bits
	}

	return nil
}

// evalBFINIT evaluates the BFINIT command responsible for initializing a
// new bloom filter and allocation it's relevant parameters based on given inputs.
// If no params are provided, it uses defaults.
func evalBFINIT(args []string) []byte {
	if len(args) != 1 && len(args) != 3 {
		return Encode(fmt.Errorf("%w for 'BFINIT' command", errWrongArgs), false)
	}

	useDefaults := false
	if len(args) == 1 {
		useDefaults = true
	}

	opts, err := newBloomOpts(args[1:], useDefaults)
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BFINIT' command", err), false)
	}

	_, err = getOrCreateBloomFilter(args[0], opts)
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BFINIT' command", err), false)
	}

	return RESP_OK
}

// evalBFADD evaluates the BFADD command responsible for adding an element to
// a bloom filter. If the filter does not exists, it will create a new one
// with default parameters.
func evalBFADD(args []string) []byte {
	if len(args) != 2 {
		return Encode(fmt.Errorf("%w for 'BFADD' command", errWrongArgs), false)
	}

	opts, _ := newBloomOpts(args[1:], true)

	bloom, err := getOrCreateBloomFilter(args[0], opts)
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BFADD' command", err), false)
	}

	resp, err := bloom.add(args[1])
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BFADD' command", err), false)
	}

	return resp
}

// evalBFEXISTS evaluates the BFEXISTS command responsible for checking existance
// of an element in a bloom filter.
func evalBFEXISTS(args []string) []byte {
	if len(args) != 2 {
		return Encode(fmt.Errorf("%w for 'BFEXISTS' command", errWrongArgs), false)
	}

	bloom, err := getOrCreateBloomFilter(args[0], nil)
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BFEXISTS' command", err), false)
	}

	resp, err := bloom.exists(args[1])
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BFEXISTS' command", err), false)
	}

	return resp
}

// evalBFINFO evaluates the BFINFO command responsible for returning the
// parameters and metadata of an existing bloom filter.
func evalBFINFO(args []string) []byte {
	if len(args) != 1 {
		return Encode(fmt.Errorf("%w for 'BFINFO' command", errWrongArgs), false)
	}

	bloom, err := getOrCreateBloomFilter(args[0], nil)
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BFINFO' command", err), false)
	}

	return Encode(bloom.info(args[0]), false)
}

// getOrCreateBloomFilter attempts to fetch an existing bloom filter from
// the kv store. If it does not exist, it tries to create one with
// given `opts` and returns it.
func getOrCreateBloomFilter(key string, opts *BloomOpts) (*Bloom, error) {
	obj := Get(key)

	// If we don't have a filter yet and `opts` are provided, create one.
	if obj == nil && opts != nil {
		obj = NewObj(newBloomFilter(opts), -1, OBJ_TYPE_BITSET, OBJ_ENCODING_BF)
		Put(key, obj)
	}

	// If no `opts` are provided for filter creation, return err
	if obj == nil && opts == nil {
		return nil, errInvalidKey
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BITSET); err != nil {
		return nil, err
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_BF); err != nil {
		return nil, err
	}

	return obj.Value.(*Bloom), nil
}
