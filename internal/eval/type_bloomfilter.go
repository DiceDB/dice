// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/object"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/twmb/murmur3"
)

const (
	defaultErrorRate float64 = 0.01
	defaultCapacity  uint64  = 1024
)

var (
	ln2      = math.Log(2)
	ln2Power = ln2 * ln2
)

var (
	errInvalidRangeErrorRateType = diceerrors.ErrGeneral("(0 < error rate range < 1) ")
	errInvalidErrorRate          = diceerrors.ErrGeneral("bad error rate")
	errInvalidCapacityType       = diceerrors.ErrGeneral("bad capacity")
	errNonPositiveCapacity       = diceerrors.ErrGeneral("(capacity should be larger than 0)")

	errEmptyValue              = diceerrors.ErrGeneral("empty value provided")
	errUnableToHash            = diceerrors.ErrGeneral("unable to hash given value")
	errInvalidInformationValue = diceerrors.ErrGeneral("Invalid information value")
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

	hashFnsSeeds []uint64 // seed for hash functions
}

type Bloom struct {
	opts   *BloomOpts // options for the bloom filter
	bitset []byte     // underlying bit representation
	cnt    uint64     // number of elements in the bloom
}

// newBloomOpts extracts the user defined values from `args`. It falls back to
// default values if `useDefaults` is set to true. Using those values, it
// creates and returns the options for bloom filter.

func defaultBloomOpts() *BloomOpts {
	return &BloomOpts{errorRate: defaultErrorRate, capacity: defaultCapacity}
}

func newBloomOpts(args []string) (*BloomOpts, error) {
	errorRate, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return nil, errInvalidErrorRate
	}

	if errorRate <= 0 || errorRate >= 1.0 {
		return nil, errInvalidRangeErrorRateType
	}

	capacity, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errInvalidCapacityType
	}

	if capacity <= 0 {
		return nil, errNonPositiveCapacity
	}

	capacityUint, _ := strconv.ParseUint(args[1], 10, 64)

	return &BloomOpts{errorRate: errorRate, capacity: capacityUint}, nil
}

// newBloomFilter creates and returns a new filter. It is responsible for initializing the
// underlying bit array.
func NewBloomFilter(opts *BloomOpts) *Bloom {
	// Calculate bits per element
	// 		bpe = -log(errorRate)/ln(2)^2
	num := -1 * math.Log(opts.errorRate)
	opts.bpe = num / ln2Power

	// Calculate the number of hash functions to be used
	// 		k = ceil(ln(2) * bpe)
	k := math.Ceil(ln2 * opts.bpe)
	opts.hashFns = make([]hash.Hash64, int(k))
	opts.hashFnsSeeds = make([]uint64, int(k))
	// Initialize hash functions with random seeds
	for i := 0; i < int(k); i++ {
		opts.hashFnsSeeds[i] = rand.Uint64() //nolint:gosec
		opts.hashFns[i] = murmur3.SeedNew64(opts.hashFnsSeeds[i])
	}

	// initialize the common slice for storing indexes of bits to be set
	opts.indexes = make([]uint64, len(opts.hashFns))

	// Calculate the number of bytes to be used
	// 		bits = k * entries / ln(2)
	//		bytes = bits * 8
	bits := uint64(math.Ceil((k * float64(opts.capacity)) / ln2))
	var bytesNeeded uint64
	if bits%8 == 0 {
		bytesNeeded = bits / 8
	} else {
		bytesNeeded = (bits / 8) + 1
	}
	opts.bits = bytesNeeded * 8

	bitset := make([]byte, bytesNeeded)

	return &Bloom{opts, bitset, 0}
}

func (b *Bloom) info(opt string) ([]interface{}, error) {
	result := make([]interface{}, 0, 10)
	if strings.EqualFold(opt, "") {
		result = append(result,
			"Capacity", b.opts.capacity,
			"Size", b.opts.bits,
			"Number of filters", len(b.opts.hashFns),
			"Number of items inserted", b.cnt,
			"Expansion rate", 2,
		)
	} else {
		switch strings.ToUpper(opt) {
		case CAPACITY:
			result = append(result, "Capacity", b.opts.capacity)
		case SIZE:
			result = append(result, "Size", b.opts.bits)
		case FILTERS:
			result = append(result, "Number of filters", len(b.opts.hashFns))
		case ITEMS:
			result = append(result, "Number of items inserted", b.cnt)
		case EXPANSION:
			result = append(result, "Expansion rate", 2)
		default:
			return nil, errInvalidInformationValue
		}
	}
	return result, nil
}

// add adds a new entry for `value` in the filter. It hashes the given
// value and sets the bit of the underlying bitset. Returns "-1" in
// case of errors, "0" if all the bits were already set and "1" if
// at least 1 new bit was set.
func (b *Bloom) add(value string) (interface{}, error) {
	// We're sure that empty values will be handled upper functions itself.
	// This is just a property check for the bloom struct.
	if value == utils.EmptyStr {
		return IntegerNegativeOne, errEmptyValue
	}

	// Update the indexes where bits are supposed to be set
	err := b.opts.updateIndexes(value)
	if err != nil {
		fmt.Println("error in getting indexes for value:", value, "err:", err)
		return IntegerNegativeOne, errUnableToHash
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
		return IntegerZero, nil
	}
	b.cnt++
	return IntegerOne, nil
}

// exists checks if the given `value` exists in the filter or not.
// It hashes the given value and checks if the bits are set or not in
// the underlying bitset. Returns "-1" in case of errors, "0" if the
// element surely does not exist in the filter, and "1" if the element
// may or may not exist in the filter.
func (b *Bloom) exists(value string) (RespType, error) {
	// We're sure that empty values will be handled upper functions itself.
	// This is just a property check for the bloom struct.
	if value == utils.EmptyStr {
		return IntegerOne, errEmptyValue
	}

	// Update the indexes where bits are supposed to be set
	err := b.opts.updateIndexes(value)
	if err != nil {
		fmt.Println("error in getting indexes for value:", value, "err:", err)
		return IntegerNegativeOne, errUnableToHash
	}

	// Check if all the bits at given indexes are set or not
	// Ideally if the element is present, we should find all set bits.
	for _, v := range b.opts.indexes {
		if !isBitSet(b.bitset, v) {
			// Return with "0" as we found one non-set bit (which is enough to conclude)
			return IntegerZero, nil
		}
	}

	// We reached here, which means the element may exist in the filter. Return "1" now.
	return IntegerOne, nil
}

// DeepCopy creates a deep copy of the Bloom struct
func (b *Bloom) DeepCopy() *Bloom {
	if b == nil {
		return nil
	}

	// Copy the BloomOpts
	copyOpts := &BloomOpts{
		errorRate: b.opts.errorRate,
		capacity:  b.opts.capacity,
		bits:      b.opts.bits,
		bpe:       b.opts.bpe,
		hashFns:   make([]hash.Hash64, len(b.opts.hashFns)),
		indexes:   make([]uint64, len(b.opts.indexes)),
	}

	// Deep copy the hash functions (assuming they are shallow copyable)
	copy(copyOpts.hashFns, b.opts.hashFns)

	// Deep copy the indexes slice
	copy(copyOpts.indexes, b.opts.indexes)

	// Deep copy the bitset
	copyBitset := make([]byte, len(b.bitset))
	copy(copyBitset, b.bitset)

	return &Bloom{
		opts:   copyOpts,
		bitset: copyBitset,
		cnt:    b.cnt,
	}
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

// CreateOrReplaceBloomFilter creates a new bloom filter with given `opts`
// and stores it in the kv store. If the bloom filter already exists, it
// replaces the existing one. If `opts` is nil, it uses the default options.
func CreateOrReplaceBloomFilter(key string, opts *BloomOpts, store *dstore.Store) *Bloom {
	if opts == nil {
		opts = defaultBloomOpts()
	}
	bf := NewBloomFilter(opts)
	obj := store.NewObj(bf, -1, object.ObjTypeBF)
	store.Put(key, obj)
	return bf
}

// GetOrCreateBloomFilter fetches an existing bloom filter from
// the kv store and returns the datastructure instance of it.
// If it does not exist, it tries to create one with given `opts` and returns it.
// Note: It also stores it in the kv store.
func GetOrCreateBloomFilter(key string, store *dstore.Store, opts *BloomOpts) (*Bloom, error) {
	bf, err := GetBloomFilter(key, store)
	if err != nil && err != diceerrors.ErrKeyNotFound {
		return nil, err
	} else if err != nil && err == diceerrors.ErrKeyNotFound {
		bf = CreateOrReplaceBloomFilter(key, opts, store)
	}
	return bf, nil
}

// GetBloomFilter fetches an existing bloom filter from
// the kv store and returns the datastructure instance of it.
// The function also returns diceerrors.ErrKeyNotFound if the key does not exist.
// It also returns diceerrors.ErrWrongTypeOperation if the object is not a bloom filter.
func GetBloomFilter(key string, store *dstore.Store) (*Bloom, error) {
	obj := store.Get(key)
	if obj == nil {
		return nil, diceerrors.ErrKeyNotFound
	}
	if err := object.AssertType(obj.Type, object.ObjTypeBF); err != nil {
		return nil, diceerrors.ErrWrongTypeOperation
	}

	return obj.Value.(*Bloom), nil
}

func (b *Bloom) Serialize(buf *bytes.Buffer) error {
	// Serialize the Bloom struct
	if err := binary.Write(buf, binary.BigEndian, b.cnt); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, b.opts.errorRate); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, b.opts.capacity); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, b.opts.bits); err != nil {
		return err
	}

	// Serialize the number of seeds and the seeds themselves
	numSeeds := uint64(len(b.opts.hashFnsSeeds))
	if err := binary.Write(buf, binary.BigEndian, numSeeds); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, b.opts.hashFnsSeeds); err != nil {
		return err
	}

	// Serialize the number of indexes and the indexes themselves
	numIndexes := uint64(len(b.opts.indexes))
	if err := binary.Write(buf, binary.BigEndian, numIndexes); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, b.opts.indexes); err != nil {
		return err
	}

	// Serialize the bitset
	if _, err := buf.Write(b.bitset); err != nil {
		return err
	}

	return nil
}

func DeserializeBloom(buf *bytes.Reader) (*Bloom, error) {
	bloom := &Bloom{
		opts: &BloomOpts{}, // Initialize the opts field to prevent nil pointer dereference
	}

	// Deserialize the Bloom struct
	if err := binary.Read(buf, binary.BigEndian, &bloom.cnt); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &bloom.opts.errorRate); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &bloom.opts.capacity); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &bloom.opts.bits); err != nil {
		return nil, err
	}

	// Deserialize hash function seeds
	var numSeeds uint64
	if err := binary.Read(buf, binary.BigEndian, &numSeeds); err != nil {
		return nil, err
	}
	bloom.opts.hashFnsSeeds = make([]uint64, numSeeds)
	if err := binary.Read(buf, binary.BigEndian, &bloom.opts.hashFnsSeeds); err != nil {
		return nil, err
	}

	// Deserialize indexes
	var numIndexes uint64
	if err := binary.Read(buf, binary.BigEndian, &numIndexes); err != nil {
		return nil, err
	}
	bloom.opts.indexes = make([]uint64, numIndexes)
	if err := binary.Read(buf, binary.BigEndian, &bloom.opts.indexes); err != nil {
		return nil, err
	}

	// Deserialize bitset
	bloom.bitset = make([]byte, bloom.opts.bits)
	if _, err := buf.Read(bloom.bitset); err != nil {
		return nil, err
	}

	// Recalculate derived values
	bloom.opts.bpe = -1 * math.Log(bloom.opts.errorRate) / math.Ln2
	bloom.opts.hashFns = make([]hash.Hash64, len(bloom.opts.hashFnsSeeds))
	for i := 0; i < len(bloom.opts.hashFnsSeeds); i++ {
		bloom.opts.hashFns[i] = murmur3.SeedNew64(bloom.opts.hashFnsSeeds[i])
	}

	return bloom, nil
}
