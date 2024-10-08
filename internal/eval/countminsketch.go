package eval

import (
	"hash"
	"hash/fnv"
	"math"
	"strconv"

	diceerrors "github.com/dicedb/dice/internal/errors"
)

type CountMinSketchOpts struct {
	depth  uint64
	width  uint64
	hasher hash.Hash64
}

type CountMinSketch struct {
	opts *CountMinSketchOpts

	count [][]uint64
}

func newCountMinSketchOpts(args []string) (*CountMinSketchOpts, error) {
	depth, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil || depth <= 0 {
		return nil, diceerrors.NewErr("invalid depth")
	}

	width, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil || width <= 0 {
		return nil, diceerrors.NewErr("invalid width")
	}

	return &CountMinSketchOpts{depth: depth, width: width, hasher: fnv.New64()}, nil
}

func newCountMinSketchOptsWithErrorRate(args []string) (*CountMinSketchOpts, error) {
	errorRate, err := strconv.ParseFloat(args[0], 10)
	if err != nil || errorRate <= 0 || errorRate >= 1.0 {
		return nil, diceerrors.NewErr("invalid overestimation value")
	}

	probability, err := strconv.ParseFloat(args[0], 10)
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

	cms.count = make([][]uint64, opts.depth)

	for row := uint64(0); row < opts.depth; row++ {
		cms.count[row] = make([]uint64, opts.width)
	}

	return cms
}
