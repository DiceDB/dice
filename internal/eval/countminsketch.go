// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"slices"
	"strconv"
	"strings"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
)

type CountMinSketchOpts struct {
	depth  uint64      // depth of the count min sketch matrix
	width  uint64      // width of the count min sketch matrix
	hasher hash.Hash64 // the hash function used to hash the key
}

// CountMinSketch implements a Count-Min Sketch as described by Cormode and
// Muthukrishnan in their paper:
// "An Improved Data Stream Summary: The Count-Min Sketch and its Applications"
// (http://dimacs.rutgers.edu/~graham/pubs/papers/cm-full.pdf).
//
// A Count-Min Sketch (CMS) is a space-efficient, probabilistic data structure
// for approximating the frequency of events in a data stream. Instead of using
// large space like a hash map, it trades accuracy for space by allowing a configurable
// error margin. Similar to Counting Bloom filters, each item is hashed into multiple
// buckets, and the item's frequency is estimated by taking the minimum count across
// those buckets.
//
// CMS is particularly useful for tracking event frequencies in large or unbounded
// data streams where storing all data or maintaining a counter for each event
// in memory is infeasible. It provides an efficient solution for real-time processing
// with minimal memory usage.
type CountMinSketch struct {
	opts *CountMinSketchOpts

	matrix [][]uint64 // the underlying matrix that stores the counts
	count  uint64     // total number of occurrences seen by the sketch
}

// newCountMinSketchOpts extracts the depth and width of the matrix when these values
// are provided by the user. depth and width must be positive integers. It returns the
// options used to create a new Count Min Sketch.
func newCountMinSketchOpts(args []string) (*CountMinSketchOpts, error) {
	width, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil || width <= 0 {
		return nil, diceerrors.NewErr("invalid width")
	}

	depth, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil || depth <= 0 {
		return nil, diceerrors.NewErr("invalid depth")
	}

	return &CountMinSketchOpts{depth: depth, width: width, hasher: fnv.New64()}, nil
}

// newCountMinSketchOptsWithErrorRate calculates the depth and width of the matrix based
// on the given permissible error rate (ε) and probability (δ). Both the values must lie
// between zero and one. A given error rate (ε) means the count estimate may exceed
// the actual count by at most ε * N, where N is the total number of elements processed,
// and the probability (1 - δ) guarantees this bound holds with at least (1 - δ) confidence.
// It returns the options used to create a new Count Min Sketch.
func newCountMinSketchOptsWithErrorRate(args []string) (*CountMinSketchOpts, error) {
	errorRate, err := strconv.ParseFloat(args[0], 64)
	if err != nil || errorRate <= 0 || errorRate >= 1.0 {
		return nil, diceerrors.NewErr("invalid overestimation value")
	}

	probability, err := strconv.ParseFloat(args[1], 64)
	if err != nil || probability <= 0 || probability >= 1.0 {
		return nil, diceerrors.NewErr("invalid prob value")
	}

	// These formulas are taken from the original paper that introduced Count Min Sketch.
	// Link to paper - http://dimacs.rutgers.edu/~graham/pubs/papers/cm-full.pdf
	width := uint64(math.Ceil(math.Exp(1) / errorRate))
	depth := uint64(math.Ceil(math.Log(1 / probability)))

	return &CountMinSketchOpts{depth: depth, width: width, hasher: fnv.New64()}, nil
}

// newCountMinSketch creates a new Count Min Sketch with given options.
// It also initializes the underlying matrix.
func newCountMinSketch(opts *CountMinSketchOpts) *CountMinSketch {
	cms := &CountMinSketch{
		opts: opts,
	}

	cms.matrix = make([][]uint64, opts.depth)
	flatMatrix := make([]uint64, opts.depth*opts.width) // single memory allocation
	for row := uint64(0); row < opts.depth; row++ {
		cms.matrix[row] = flatMatrix[row*opts.width : (row+1)*opts.width : (row+1)*opts.width]
	}

	return cms
}

// returns information about the underlying matrix for the given Count Min Sketch.
func (c *CountMinSketch) info() []interface{} {
	info := make([]interface{}, 0, 3)

	info = append(info, "width", c.opts.width, "depth", c.opts.depth, "count", c.count)

	return info
}

// this function computes the base hash values which are then used to generate
// other hash values for the given key.
func (c *CountMinSketch) baseHashes(key []byte) (hash1, hash2 uint32) {
	c.opts.hasher.Reset()
	c.opts.hasher.Write(key)

	sum := c.opts.hasher.Sum(nil)

	upper := sum[0:4]
	lower := sum[4:8]

	hash1 = binary.BigEndian.Uint32(upper)
	hash2 = binary.BigEndian.Uint32(lower)

	return
}

// returns the positions in the matrix where the count of the given key
// should be updated.
func (c *CountMinSketch) matrixPositions(key []byte) (positions []uint64) {
	positions = make([]uint64, c.opts.depth)

	hash1, hash2 := c.baseHashes(key)

	uintHash1 := uint64(hash1)
	uintHash2 := uint64(hash2)

	for row := uint64(0); row < c.opts.depth; row++ {
		positions[row] = (uintHash1 + uintHash2*row) % c.opts.width
	}
	return
}

// updates the underlying matrix for the given key by count.
func (c *CountMinSketch) updateMatrix(key string, count uint64) {
	for row, col := range c.matrixPositions([]byte(key)) {
		c.matrix[row][col] += count
	}
	c.count += count
}

// estimateCount is used to query the sketch for the value of a key.
// The estimated count is the minimum of the values present at the
// positons for a given key.
func (c *CountMinSketch) estimateCount(key string) uint64 {
	var count uint64 = math.MaxUint64
	for row, col := range c.matrixPositions([]byte(key)) {
		if c.matrix[row][col] < count {
			count = c.matrix[row][col]
		}
	}

	return count
}

// returns a deep copy of the Count Min Sketch
func (c *CountMinSketch) DeepCopy() *CountMinSketch {
	if c == nil {
		return nil
	}

	copyOpts := &CountMinSketchOpts{
		depth:  c.opts.depth,
		width:  c.opts.width,
		hasher: c.opts.hasher,
	}

	// Deep copy the matrix
	matrix := make([][]uint64, c.opts.depth)
	flatMatrix := make([]uint64, c.opts.depth*c.opts.width) // single memory allocation
	for row := uint64(0); row < c.opts.depth; row++ {
		matrix[row] = flatMatrix[row*c.opts.width : (row+1)*c.opts.width : (row+1)*c.opts.width]
		copy(matrix[row], c.matrix[row])
	}

	return &CountMinSketch{
		opts:   copyOpts,
		matrix: matrix,
		count:  c.count,
	}
}

// mergeMatrices combines two or more Count Min Sketches and puts the result in another Count Min Sketch.
// The merging is done based on the weights assigned to each sketch. The counts stored in the original sketch
// are ignored. The array of source sketches might include the destination sketch too. That is handled by
// making a deep copy of the destination sketch.
func (c *CountMinSketch) mergeMatrices(sources []*CountMinSketch, weights []uint64, originalKey string, keys []string) {
	originalCopy := c.DeepCopy()

	// resets the destination sketch
	for row := uint64(0); row < c.opts.depth; row++ {
		for col := uint64(0); col < c.opts.width; col++ {
			c.matrix[row][col] = 0
		}
	}

	// for every row and column, take the weighted sum of the source sketches.
	for row := uint64(0); row < c.opts.depth; row++ {
		for col := uint64(0); col < c.opts.width; col++ {
			for i, cms := range sources {
				if keys[i] == originalKey {
					// use the deep copy of the destination
					c.matrix[row][col] += weights[i] * originalCopy.matrix[row][col]
				} else {
					c.matrix[row][col] += weights[i] * cms.matrix[row][col]
				}
			}
		}
	}

	// update the count attribute
	c.count = 0
	for i, cms := range sources {
		if keys[i] == originalKey {
			c.count += weights[i] * originalCopy.count
		} else {
			c.count += weights[i] * cms.count
		}
	}
}

// serialize encodes the CountMinSketch into a byte slice.
func (c *CountMinSketch) serialize(buffer *bytes.Buffer) error {
	if c == nil {
		return errors.New("cannot serialize a nil CountMinSketch")
	}

	// Write depth, width, and count
	if err := binary.Write(buffer, binary.BigEndian, c.opts.depth); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, c.opts.width); err != nil {
		return err
	}
	if err := binary.Write(buffer, binary.BigEndian, c.count); err != nil {
		return err
	}

	// Write matrix
	for i := 0; i < len(c.matrix); i++ {
		for j := 0; j < len(c.matrix[i]); j++ {
			if err := binary.Write(buffer, binary.BigEndian, c.matrix[i][j]); err != nil {
				return err
			}
		}
	}

	return nil
}

// deserialize reconstructs a CountMinSketch from a byte slice.
func DeserializeCMS(buffer *bytes.Reader) (*CountMinSketch, error) {
	if buffer.Len() < 24 { // Minimum size for depth, width, and count
		return nil, errors.New("insufficient data for deserialization")
	}

	var depth, width, count uint64

	// Read depth, width, and count
	if err := binary.Read(buffer, binary.BigEndian, &depth); err != nil {
		return nil, err
	}
	if err := binary.Read(buffer, binary.BigEndian, &width); err != nil {
		return nil, err
	}
	if err := binary.Read(buffer, binary.BigEndian, &count); err != nil {
		return nil, err
	}
	// fmt.Println(depth, width, count, buffer.Len())
	// Validate data size
	expectedSize := int(depth * width * 8) // Each uint64 takes 8 bytes
	if buffer.Len() <= expectedSize {
		return nil, errors.New("data size mismatch with expected matrix size")
	}

	// Read matrix
	matrix := make([][]uint64, depth)
	flatMatrix := make([]uint64, depth*width) // single memory allocation
	for i := 0; i < int(depth); i++ {
		matrix[i] = flatMatrix[i*int(width) : (i+1)*int(width) : (i+1)*int(width)]
		for j := 0; j < int(width); j++ {
			if err := binary.Read(buffer, binary.BigEndian, &matrix[i][j]); err != nil {
				return nil, err
			}
		}
	}

	opts := &CountMinSketchOpts{
		depth:  depth,
		width:  width,
		hasher: fnv.New64(), // Default hasher
	}

	return &CountMinSketch{
		opts:   opts,
		matrix: matrix,
		count:  count,
	}, nil
}

// evalCMSMerge is used to merge multiple sketches into one. The final sketch
// contains the weighted sum of the values in each of the source sketches. If
// weights are not provided, default is 1.
func evalCMSMerge(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("cms.merge"),
		}
	}

	destination, err := getCountMinSketch(args[0], store)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.merge' command", err)),
		}
	}

	numberOfKeys, err := strconv.ParseInt(args[1], 10, 64)

	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("cannot parse number for 'cms.merge' command"),
		}
	}

	if numberOfKeys < 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("invalid number of arguments to merge for 'cms.merge' command"),
		}
	}

	if len(args) != int(2+numberOfKeys) && len(args) != int(3+2*numberOfKeys) {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("invalid number of arguments to merge for 'cms.merge' command"),
		}
	}
	keys := args[2 : 2+numberOfKeys]
	sources := make([]*CountMinSketch, 0, numberOfKeys)

	for _, key := range keys {
		c, err := getCountMinSketch(key, store)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.merge' command", err)),
			}
		}
		if c.opts.depth != destination.opts.depth || c.opts.width != destination.opts.width {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("width/depth doesn't match"),
			}
		}
		sources = append(sources, c)
	}

	if len(args) == int(2+numberOfKeys) {
		weights := slices.Repeat([]uint64{1}, int(numberOfKeys))
		destination.mergeMatrices(sources, weights, args[0], keys)

		return &EvalResponse{
			Result: OK,
			Error:  nil,
		}
	}

	if !strings.EqualFold(args[2+numberOfKeys], "WEIGHTS") {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("invalid number of arguments to merge for 'cms.merge' command"),
		}
	}

	numberOfWeights := len(args) - 3 - int(numberOfKeys)
	if int(numberOfKeys) != numberOfWeights {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral("invalid number of arguments to merge for 'cms.merge' command"),
		}
	}

	values := args[3+numberOfWeights:]
	weights := make([]uint64, 0, numberOfWeights)
	for _, value := range values {
		weight, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid number"),
			}
		}
		weights = append(weights, weight)
	}

	destination.mergeMatrices(sources, weights, args[0], keys)

	return &EvalResponse{
		Result: OK,
		Error:  nil,
	}
}

// evalCMSQuery returns the count for one or more items in a sketch.
func evalCMSQuery(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 2 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("cms.query"),
		}
	}

	cms, err := getCountMinSketch(args[0], store)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.query' command", err)),
		}
	}

	results := make([]uint64, 0, len(args[1:]))

	for _, key := range args[1:] {
		results = append(results, cms.estimateCount(key))
	}

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

// evalCMSIncrBy increases the count of item by increment. Multiple items can be increased with one call.
func evalCMSIncrBy(args []string, store *dstore.Store) *EvalResponse {
	if len(args) < 3 || len(args)%2 == 0 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("cms.incrby"),
		}
	}

	cms, err := getCountMinSketch(args[0], store)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.incrby' command", err)),
		}
	}

	keyValuePairs := args[1:]
	for iter := 1; iter < len(keyValuePairs); iter += 2 {
		_, err := strconv.ParseUint(keyValuePairs[iter], 10, 64)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("cannot parse number for 'cms.incrby' command"),
			}
		}
	}

	results := make([]uint64, 0, len(keyValuePairs)/2)

	for iter := 0; iter <= len(keyValuePairs)-2; iter += 2 {
		key := keyValuePairs[iter]
		value, err := strconv.ParseUint(keyValuePairs[iter+1], 10, 64)
		if err != nil {
			return &EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("cannot parse number"),
			}
		}

		cms.updateMatrix(key, value)
		count := cms.estimateCount(key)
		results = append(results, count)
	}

	return &EvalResponse{
		Result: results,
		Error:  nil,
	}
}

// evalCMSINFO returns width, depth and total count of the sketch.
func evalCMSINFO(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 1 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("cms.info"),
		}
	}
	cms, err := getCountMinSketch(args[0], store)
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.info' command", err)),
		}
	}

	return &EvalResponse{
		Result: cms.info(),
		Error:  nil,
	}
}

// evalCMSINITBYDIM initializes a Count-Min Sketch by dimensions (width and depth) specified in the call.
func evalCMSINITBYDIM(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("cms.initbydim"),
		}
	}

	opts, err := newCountMinSketchOpts(args[1:])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.initbydim' command", err)),
		}
	}

	if err = createCountMinSketch(args[0], opts, store); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.initbydim' command", err)),
		}
	}

	return &EvalResponse{
		Result: OK,
		Error:  nil,
	}
}

// evalCMSINITBYPROB initializes a Count-Min Sketch for a given error rate and probability.
// Error rate is used to calculate the width while probability is used to calculate the depth
// of the sketch.
func evalCMSINITBYPROB(args []string, store *dstore.Store) *EvalResponse {
	if len(args) != 3 {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrWrongArgumentCount("cms.initbyprob"),
		}
	}

	opts, err := newCountMinSketchOptsWithErrorRate(args[1:])
	if err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.initbyprob' command", err)),
		}
	}

	if err = createCountMinSketch(args[0], opts, store); err != nil {
		return &EvalResponse{
			Result: nil,
			Error:  diceerrors.ErrGeneral(fmt.Sprintf("%v for 'cms.initbyprob' command", err)),
		}
	}

	return &EvalResponse{
		Result: OK,
		Error:  nil,
	}
}

// creates a new Count Min Sketch in the key-value store. Returns error if the key already exists.
func createCountMinSketch(key string, opts *CountMinSketchOpts, store *dstore.Store) error {
	obj := store.Get(key)

	if obj != nil {
		return diceerrors.NewErr("key already exists")
	}

	obj = store.NewObj(newCountMinSketch(opts), -1, object.ObjTypeCountMinSketch)
	store.Put(key, obj)

	return nil
}

// fetches the Count Min Sketch for the given key from the key-value store. Returns error if key
// does not exist or the key has the wrong encoding.
func getCountMinSketch(key string, store *dstore.Store) (*CountMinSketch, error) {
	obj := store.Get(key)

	if obj == nil {
		return nil, diceerrors.NewErr("key does not exist")
	}

	if err := object.AssertTypeWithError(obj.Type, object.ObjTypeCountMinSketch); err != nil {
		return nil, err
	}

	return obj.Value.(*CountMinSketch), nil
}
