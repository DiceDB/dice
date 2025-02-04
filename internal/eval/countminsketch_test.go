// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"testing"

	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

func TestCountMinSketch(t *testing.T) {
	store := dstore.NewStore(nil, nil)

	testCMSInitByDim(t, store)
	testCMSInitByProb(t, store)
	testCMSInfo(t, store)
	testCMSIncrBy(t, store)
	testCMSQuery(t, store)
	testCMSMerge(t, store)
}

func testCMSInitByDim(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"cms initbydim - wrong number of arguments": {
			input: []string{"cms_key"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("cms.initbydim"),
			},
		},
		"cms initbydim - wrong type of width": {
			input: []string{"cms_key", "not_a_number", "5"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid width for 'cms.initbydim' command"),
			},
		},
		"cms initbydim - wrong type of depth": {
			input: []string{"cms_key", "5", "not_a_number"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid depth for 'cms.initbydim' command"),
			},
		},
		"cms initbydim - negative width": {
			input: []string{"cms_key", "-100", "5"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid width for 'cms.initbydim' command"),
			},
		},
		"cms initbydim - negative depth": {
			input: []string{"cms_key", "5", "-100"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid depth for 'cms.initbydim' command"),
			},
		},

		"cms initbydim - correct width and depth": {
			input: []string{"cms_key", "1000", "5"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
		"cms initbydim - key already exists": {
			setup: func() {
				key := "cms_key"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key", "1000", "5"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("key already exists for 'cms.initbydim' command"),
			},
		},
	}

	runMigratedEvalTests(t, tests, evalCMSINITBYDIM, store)
}

func testCMSInitByProb(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"cms initbyprob - wrong number of arguments": {
			input: []string{"cms_key1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("cms.initbyprob"),
			},
		},
		"cms initbyprob - wrong type of error rate": {
			input: []string{"cms_key1", "not_a_number", "0.01"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid overestimation value for 'cms.initbyprob' command"),
			},
		},
		"cms initbyprob - wrong type of probability": {
			input: []string{"cms_key1", "0.01", "not_a_number"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid prob value for 'cms.initbyprob' command"),
			},
		},
		"cms initbyprob - error rate out of range": {
			input: []string{"cms_key1", "1", "0.01"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid overestimation value for 'cms.initbyprob' command"),
			},
		},
		"cms initbyprob - probability out of range": {
			input: []string{"cms_key1", "0.01", "1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid prob value for 'cms.initbyprob' command"),
			},
		},
		"cms initbyprob - correct error rate and probability": {
			input: []string{"cms_key1", "0.01", "0.01"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
		"cms initbyprob - key already exists": {
			setup: func() {
				key := "cms_key1"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key1", "0.01", "0.01"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("key already exists for 'cms.initbyprob' command"),
			},
		},
	}

	runMigratedEvalTests(t, tests, evalCMSINITBYPROB, store)
}

func testCMSInfo(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"cms info - wrong number of arguments": {
			input: []string{},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("cms.info"),
			},
		},
		"cms info - key doesn't exist": {
			input: []string{"cms_key2"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("key does not exist for 'cms.info' command"),
			},
		},
		"cms info - one argument": {
			setup: func() {
				key := "cms_key2"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key2"},
			migratedOutput: EvalResponse{
				Result: []interface{}{"width", uint64(1000), "depth", uint64(5), "count", uint64(0)},
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalCMSINFO, store)
}

func testCMSIncrBy(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"cms incrby - wrong number of arguments": {
			input: []string{"cms_key3"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("cms.incrby"),
			},
		},
		"cms incrby - key doesn't exist": {
			input: []string{"cms_key3", "test", "10"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("key does not exist for 'cms.incrby' command"),
			},
		},
		"cms incrby - inserting keys": {
			setup: func() {
				key := "cms_key3"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key3", "test", "10", "test1", "10"},
			migratedOutput: EvalResponse{
				Result: []uint64{10, 10},
				Error:  nil,
			},
		},
		"cms incrby - missing values": {
			setup: func() {
				key := "cms_key3"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key3", "test", "10", "test1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("cms.incrby"),
			},
		},
		"cms incrby - negative values": {
			setup: func() {
				key := "cms_key3"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key3", "test", "-1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("cannot parse number for 'cms.incrby' command"),
			},
		},
	}

	runMigratedEvalTests(t, tests, evalCMSIncrBy, store)
}

func testCMSQuery(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"cms query - wrong number of arguments": {
			input: []string{"cms_key4"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("cms.query"),
			},
		},
		"cms query - key doesn't exist": {
			input: []string{"cms_key4", "test"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("key does not exist for 'cms.query' command"),
			},
		},
		"cms query - query keys": {
			setup: func() {
				key := "cms_key4"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
				cms, _ := getCountMinSketch(key, store)

				cms.updateMatrix("test", 10000)
				cms.updateMatrix("test1", 100)
			},
			input: []string{"cms_key4", "test", "test1"},
			migratedOutput: EvalResponse{
				Result: []uint64{10000, 100},
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalCMSQuery, store)
}

func testCMSMerge(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"cms merge - wrong number of arguments": {
			input: []string{"cms_key5"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrWrongArgumentCount("cms.merge"),
			},
		},
		"cms merge - key doesn't exist": {
			input: []string{"cms_key5", "1", "test"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("key does not exist for 'cms.merge' command"),
			},
		},
		"cms merge - wrong type of number of sources": {
			setup: func() {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

			},
			input: []string{"cms_key5", "not_a_number", "test"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("cannot parse number for 'cms.merge' command"),
			},
		},
		"cms merge - more sources than specified": {
			setup: func() {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

				source2 := "test1"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source2, opts, store)

			},
			input: []string{"cms_key5", "3", "test", "test1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid number of arguments to merge for 'cms.merge' command"),
			},
		},
		"cms merge - fewer sources than specified": {
			setup: func() {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

				source2 := "test1"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source2, opts, store)

			},
			input: []string{"cms_key5", "1", "test", "test1"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid number of arguments to merge for 'cms.merge' command"),
			},
		},
		"cms merge - missing weights": {
			setup: func() {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

			},
			input: []string{"cms_key5", "1", "test", "WEIGHTS"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid number of arguments to merge for 'cms.merge' command"),
			},
		},
		"cms merge - more weights than needed": {
			setup: func() {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

			},
			input: []string{"cms_key5", "1", "test", "WEIGHTS", "1", "2"},
			migratedOutput: EvalResponse{
				Result: nil,
				Error:  diceerrors.ErrGeneral("invalid number of arguments to merge for 'cms.merge' command"),
			},
		},
		"cms merge - correct case": {
			setup: func() {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

			},
			input: []string{"cms_key5", "1", "test"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
		"cms merge - correct case with given weights": {
			setup: func() {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

				source2 := "test1"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source2, opts, store)
			},
			input: []string{"cms_key5", "2", "test", "test1", "WEIGHTS", "1", "2"},
			migratedOutput: EvalResponse{
				Result: OK,
				Error:  nil,
			},
		},
	}

	runMigratedEvalTests(t, tests, evalCMSMerge, store)
}
