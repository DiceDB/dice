package eval

import (
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"strconv"

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

	return info
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

	if _, err = createCountMinSketch(args[0], opts, store); err != nil {
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

	_, err = createCountMinSketch(args[0], opts, store)
	if err != nil {
		return diceerrors.NewErrWithFormattedMessage("%w for 'CMS.INITBYPROB' command", err)
	}

	return clientio.RespOK
}

func createCountMinSketch(key string, opts *CountMinSketchOpts, store *dstore.Store) (*CountMinSketch, error) {
	obj := store.Get(key)

	if obj != nil {
		return nil, diceerrors.NewErr("key already exists")
	}

	obj = store.NewObj(newCountMinSketch(opts), -1, object.ObjTypeCountMinSketch, object.ObjEncodingMatrix)
	store.Put(key, obj)

	return obj.Value.(*CountMinSketch), nil
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
