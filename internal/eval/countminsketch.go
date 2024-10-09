package eval

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
)

type CountMinSketchOpts struct {
	depth  uint64
	width  uint64
	hasher hash.Hash64
}

type CountMinSketch struct {
	opts *CountMinSketchOpts

	matrix [][]uint64
	count  uint64
}

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

func newCountMinSketchOptsWithErrorRate(args []string) (*CountMinSketchOpts, error) {
	errorRate, err := strconv.ParseFloat(args[0], 64)
	if err != nil || errorRate <= 0 || errorRate >= 1.0 {
		return nil, diceerrors.NewErr("invalid overestimation value")
	}

	probability, err := strconv.ParseFloat(args[0], 64)
	if err != nil || probability <= 0 || probability >= 1.0 {
		return nil, diceerrors.NewErr("invalid overestimation value")
	}

	width := uint64(math.Ceil(math.Exp(1) / errorRate))
	depth := uint64(math.Ceil(math.Log(1 / probability)))

	return &CountMinSketchOpts{depth: depth, width: width, hasher: fnv.New64()}, nil
}

func newCountMinSketch(opts *CountMinSketchOpts) *CountMinSketch {
	cms := &CountMinSketch{
		opts: opts,
	}

	cms.matrix = make([][]uint64, opts.depth)

	for row := uint64(0); row < opts.depth; row++ {
		cms.matrix[row] = make([]uint64, opts.width)
	}

	return cms
}

func (c *CountMinSketch) info(name string) string {
	info := utils.EmptyStr
	if name != utils.EmptyStr {
		info = "name: " + name + ", "
	}
	info += fmt.Sprintf("width: %d, ", c.opts.width)
	info += fmt.Sprintf("depth: %d,", c.opts.depth)
	info += fmt.Sprintf("count: %d,", c.count)

	return info
}

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

func (c *CountMinSketch) updateMatrix(key string, count uint64) {
	for row, col := range c.matrixPositions([]byte(key)) {
		c.matrix[row][col] += count
	}
	c.count += count
}

func (c *CountMinSketch) estimateCount(key string) uint64 {
	var count uint64 = math.MaxUint64
	for row, col := range c.matrixPositions([]byte(key)) {
		if c.matrix[row][col] < count {
			count = c.matrix[row][col]
		}
	}

	return count
}

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
	for row := uint64(0); row < c.opts.depth; row++ {
		matrix[row] = make([]uint64, c.opts.width)
		copy(matrix[row], c.matrix[row])
	}

	return &CountMinSketch{
		opts:   copyOpts,
		matrix: matrix,
		count:  c.count,
	}
}

func (c *CountMinSketch) mergeMatrices(sources []*CountMinSketch, weights []uint64, originalKey string, keys []string) {
	originalCopy := c.DeepCopy()

	for row := uint64(0); row < c.opts.depth; row++ {
		for col := uint64(0); col < c.opts.width; col++ {
			c.matrix[row][col] = 0
		}
	}

	for row := uint64(0); row < c.opts.depth; row++ {
		for col := uint64(0); col < c.opts.width; col++ {
			for i, cms := range sources {
				if keys[i] == originalKey {
					c.matrix[row][col] += weights[i] * originalCopy.matrix[row][col]
				} else {
					c.matrix[row][col] += weights[i] * cms.matrix[row][col]
				}
			}
		}
	}

	c.count = 0
	for i, cms := range sources {
		if keys[i] == originalKey {
			c.count += weights[i] * originalCopy.count
		} else {
			c.count += weights[i] * cms.count
		}
	}
}

func evalCMSMerge(args []string, store *dstore.Store) []byte {
	if len(args) < 3 {
		return diceerrors.NewErrArity("CMS.MERGE")
	}

	destination, err := getCountMinSketch(args[0], store)
	if err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.MERGE' command", err)
	}

	numberOfKeys, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || numberOfKeys < 1 {
		return diceerrors.NewErrWithMessage("invalid number of keys to merge")
	}

	keys := args[2 : 2+numberOfKeys]
	sources := make([]*CountMinSketch, 0, numberOfKeys)

	for _, key := range keys {
		c, err := getCountMinSketch(key, store)
		if err != nil {
			return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.MERGE' command", err)
		}
		if c.opts.depth != destination.opts.depth || c.opts.width != destination.opts.width {
			return diceerrors.NewErrWithMessage("width/depth doesn't match")
		}
		sources = append(sources, c)
	}

	if len(args) == int(2+numberOfKeys) {
		weights := slices.Repeat([]uint64{1}, int(numberOfKeys))
		destination.mergeMatrices(sources, weights, args[0], keys)

		return clientio.RespOK
	}

	if !strings.EqualFold(args[2+numberOfKeys], "WEIGHTS") {
		return diceerrors.NewErrWithMessage("syntax error")
	}

	numberOfWeights := len(args) - 3 - int(numberOfKeys)
	if int(numberOfKeys) != numberOfWeights {
		return diceerrors.NewErrWithMessage("syntax error")
	}

	values := args[3+numberOfWeights:]
	weights := make([]uint64, 0, numberOfWeights)
	for _, value := range values {
		weight, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return diceerrors.NewErrWithMessage("invalid number")
		}
		weights = append(weights, weight)
	}

	destination.mergeMatrices(sources, weights, args[0], keys)

	return clientio.RespOK
}

func evalCMSQuery(args []string, store *dstore.Store) []byte {
	if len(args) < 2 {
		return diceerrors.NewErrArity("CMS.QUERY")
	}

	cms, err := getCountMinSketch(args[0], store)
	if err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.QUERY' command", err)
	}

	results := make([]uint64, 0, len(args[1:]))

	for _, key := range args[1:] {
		results = append(results, cms.estimateCount(key))
	}

	return clientio.Encode(results, false)
}

func evalCMSIncrBy(args []string, store *dstore.Store) []byte {
	if len(args) < 3 || len(args)%2 == 0 {
		return diceerrors.NewErrArity("CMS.INCRBY")
	}

	cms, err := getCountMinSketch(args[0], store)
	if err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.INCRBY' command", err)
	}

	keyValuePairs := args[1:]
	for iter := 1; iter < len(keyValuePairs); iter += 2 {
		_, err := strconv.ParseUint(keyValuePairs[iter], 10, 64)
		if err != nil {
			return diceerrors.NewErrWithMessage("cannot parse number")
		}
	}

	results := make([]uint64, 0, len(keyValuePairs)/2)

	for iter := 0; iter <= len(keyValuePairs)-2; iter += 2 {
		key := keyValuePairs[iter]
		value, err := strconv.ParseUint(keyValuePairs[iter+1], 10, 64)
		if err != nil {
			return diceerrors.NewErrWithMessage("cannot parse number")
		}

		cms.updateMatrix(key, value)
		count := cms.estimateCount(key)
		results = append(results, count)
	}

	return clientio.Encode(results, false)
}

func evalCMSINFO(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("CMS.INFO")
	}
	cms, err := getCountMinSketch(args[0], store)
	if err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.INFO' command", err)
	}

	return clientio.Encode(cms.info(args[0]), false)
}

func evalCMSINITBYDIM(args []string, store *dstore.Store) []byte {
	if len(args) != 3 {
		return diceerrors.NewErrArity("CMS.INITBYDIM")
	}

	opts, err := newCountMinSketchOpts(args[1:])
	if err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.INITBYDIM' command", err)
	}

	if err = createCountMinSketch(args[0], opts, store); err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.INITBYDIM' command", err)
	}

	return clientio.RespOK
}

func evalCMSINITBYPROB(args []string, store *dstore.Store) []byte {
	if len(args) != 3 {
		return diceerrors.NewErrArity("CMS.INITBYPROB")
	}

	opts, err := newCountMinSketchOptsWithErrorRate(args[1:])
	if err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.INITBYPROB' command", err)
	}

	if err = createCountMinSketch(args[0], opts, store); err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.INITBYPROB' command", err)
	}

	return clientio.RespOK
}

func createCountMinSketch(key string, opts *CountMinSketchOpts, store *dstore.Store) error {
	obj := store.Get(key)

	if obj != nil {
		return diceerrors.NewErr("key already exists")
	}

	obj = store.NewObj(newCountMinSketch(opts), -1, object.ObjTypeCountMinSketch, object.ObjEncodingMatrix)
	store.Put(key, obj)

	return nil
}

func getCountMinSketch(key string, store *dstore.Store) (*CountMinSketch, error) {
	obj := store.Get(key)

	if obj == nil {
		return nil, diceerrors.NewErr("key does not exist")
	}

	if err := object.AssertType(obj.TypeEncoding, object.ObjTypeCountMinSketch); err != nil {
		return nil, err
	}

	if err := object.AssertEncoding(obj.TypeEncoding, object.ObjEncodingMatrix); err != nil {
		return nil, err
	}

	return obj.Value.(*CountMinSketch), nil
}
