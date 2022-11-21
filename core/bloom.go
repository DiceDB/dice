package core

import (
	"errors"
	"fmt"
	"hash"
	"math"
	"strconv"
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
)

type BloomOpts struct {
	errorRate float64 // desired error rate (the false positive rate) of the filter
	capacity  uint64  // number of expected entries to be added to the filter

	bits    uint64        // total number of bits reserved for the filter
	hashFns []hash.Hash32 // array of hash functions
	bpe     float64       // bits per element
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
	opts.bpe = num / ln2Power // (TODO): type conversion or use ceil here

	// Calculate the number of hash functions to be used
	// 		k = ceil(ln(2) * bpe)
	k := math.Ceil(ln2 * opts.bpe)
	opts.hashFns = make([]hash.Hash32, int(k))

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

func evalBFInit(args []string) []byte {
	if len(args) != 3 {
		return Encode(fmt.Errorf("%w for 'BINIT' command", errWrongArgs), false)
	}

	var key string = args[0]
	opts, err := newBloomOpts(args[1:], true)
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BINIT' command", err), false)
	}

	_, err = getOrCreateBloomFilter(key, opts)
	if err != nil {
		return Encode(fmt.Errorf("%w for 'BINIT' command", err), false)
	}

	return RESP_OK
}

func evalBFAdd(args []string) []byte {
	return nil
}

func evalBFExists(args []string) []byte {
	return nil
}

func evalBFInfo(args []string) []byte {
	return nil
}

// getOrCreateBloomFilter attempts to fetch an existing bloom filter from
// the kv store. If it does not exist, it tries to create one with
// given `opts` and returns it.
func getOrCreateBloomFilter(key string, opts *BloomOpts) (*Bloom, error) {
	obj := Get(key)
	if obj == nil {
		obj = NewObj(newBloomFilter(opts), -1, OBJ_TYPE_BITSET, OBJ_ENCODING_BF)
		Put(key, obj)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_BITSET); err != nil {
		return nil, err
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_BF); err != nil {
		return nil, err
	}

	return obj.Value.(*Bloom), nil
}
