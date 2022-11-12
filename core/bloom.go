package core

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	defaultErrorRate float64 = 0.01 // (TODO): Update values
	defaultCapacity  uint64  = 1024 // (TODO): Update values
)

var (
	errWrongArgs            = errors.New("ERR wrong number of arguments")
	errInvalidErrorRateType = errors.New("ERR only float values can be provided for error rate")
	errInvalidErrorRate     = errors.New("ERR invalid error rate provided")
	errInvalidEntriessType  = errors.New("ERR only integer values can be provided for entries")
	errInvalidEntries       = errors.New("ERR invalid entries provided")
)

type BloomOpts struct {
	errorRate float64 // error rate for the bloom filter (the false positive rate)
	entries   uint64  // max number of expected entries in a bloom filter
	// add other params here
}

type Bloom struct {
	opts   *BloomOpts // options for the bloom filter
	bitset []byte     // underlying bit representation
}

// newBloomOpts creates options for a bloom filter given user defined values. It also
// falls back to default values if `useDefaults` is set to true.
func newBloomOpts(args []string, useDefaults bool) (*BloomOpts, error) {
	if useDefaults {
		return &BloomOpts{defaultErrorRate, defaultCapacity}, nil
	}

	errorRate, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return nil, errInvalidErrorRateType
	}

	if errorRate <= 0 || errorRate >= 1.0 {
		return nil, errInvalidErrorRate
	}

	entries, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return nil, errInvalidEntriessType
	}

	if entries < 1 {
		return nil, errInvalidEntries
	}

	return &BloomOpts{errorRate, entries}, nil
}

// newBloomFilter creates and returns a new filter. It is responsible for initializing the
// underlying bit array.
func newBloomFilter(opts *BloomOpts) *Bloom {
	// Allocate bit capacity here and perform other calculations if required
	return &Bloom{opts, nil}
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

	getOrCreateBloomFilter(key, opts)

	return nil
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
