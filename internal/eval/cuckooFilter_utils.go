package eval

import (
	"encoding/binary"

	"math/rand"

	"github.com/twmb/murmur3"
)

func fastrandom(n uint32) uint32 {
	return rand.Uint32() % n
}

func randomIdx(i1, i2 uint) uint {
	if fastrandom(2) == 0 {
		return i1
	}
	return i2
}

func getNextPow2(n uint64) uint64 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++

	return n
}

func getPrimaryIndexAndFingerprint(data []byte, bucketIndexMask uint) (uint, fingerPrint) {
	// Use twmb's murmur3 with a seed
	seed := uint64(1069)
	hashedVal := murmur3.SeedSum64(seed, data)
	idx1 := uint(hashedVal) & bucketIndexMask

	// Generate fingerprint
	shifted := hashedVal >> (64 - fingerPrintSizeBits)
	fp := fingerPrint(shifted%(maxFingerPrint-1) + 1)

	return idx1, fp
}

func getSecondaryIndex(fp fingerPrint, i uint, bucketIndexMask uint) uint {
	// Convert the fingerprint to a byte slice for hashing
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(fp))

	// Hash the fingerprint byte slice with twmb murmur3 and the same seed
	seed := uint64(1069)
	hashedVal := murmur3.SeedSum64(seed, b)

	return (i ^ uint(hashedVal)) & bucketIndexMask
}

// ### BUCKET OPERATIONS
func (b *bucket) contains(fp fingerPrint) bool {
	for _, val := range *b {
		if val == fp {
			return true
		}
	}
	return false
}

func (b *bucket) insert(fp fingerPrint) bool {

	for i, val := range *b {

		if val == nullFingerPrint {
			(*b)[i] = fp
			return true
		}
	}

	return false
}

func (b *bucket) delete(fp fingerPrint) bool {
	for i, val := range *b {
		if val == fp {
			(*b)[i] = nullFingerPrint
			return true
		}
	}
	return false
}

func (b *bucket) reset() {
	for i := range *b {
		(*b)[i] = nullFingerPrint
	}
}

// ### CF OPERATIONS

func (cf *CuckooFilter) add(data []byte) bool {
	primaryIndex, fp := getPrimaryIndexAndFingerprint(data, cf.bucketIndexMask)

	if cf.insert(fp, primaryIndex) {
		return true
	}

	secondaryIndex := getSecondaryIndex(fp, primaryIndex, cf.bucketIndexMask)

	if cf.insert(fp, secondaryIndex) {

		return true
	}

	return cf.reinsert(fp, randomIdx(primaryIndex, secondaryIndex))

}

func (cf *CuckooFilter) lookup(data []byte) bool {
	primaryIndex, fp := getPrimaryIndexAndFingerprint(data, cf.bucketIndexMask)

	if b := cf.buckets[primaryIndex]; b.contains(fp) {
		return true
	}

	secondaryIndex := getSecondaryIndex(fp, primaryIndex, cf.bucketIndexMask)

	b := cf.buckets[secondaryIndex]

	return b.contains(fp)
}

func (cf *CuckooFilter) remove(data []byte) bool {
	primaryIndex, fp := getPrimaryIndexAndFingerprint(data, cf.bucketIndexMask)

	secondaryIndex := getSecondaryIndex(fp, primaryIndex, cf.bucketIndexMask)

	return cf.delete(fp, primaryIndex) || cf.delete(fp, secondaryIndex)
}

func (cf *CuckooFilter) insert(fp fingerPrint, index uint) bool {

	if cf.buckets[index].insert(fp) {
		cf.count++
		return true
	}
	return false
}

func (cf *CuckooFilter) reinsert(fp fingerPrint, index uint) bool {

	for k := 0; k < maxDisplacements; k++ {
		j := fastrandom(uint32(cf.opts.bucketSize))
		cf.buckets[index][j], fp = fp, cf.buckets[index][j]
		index = getSecondaryIndex(fp, index, cf.bucketIndexMask)
		if cf.insert(fp, index) {
			return true
		}
	}

	return false
}

func (cf *CuckooFilter) delete(fp fingerPrint, index uint) bool {
	if cf.buckets[index].delete(fp) {
		cf.count--
		return true
	}
	return false
}
