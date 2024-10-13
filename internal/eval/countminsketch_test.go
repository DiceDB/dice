package eval

import (
	"testing"

	"github.com/dicedb/dice/internal/clientio"
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
		"wrong number of arguments": {
			input: []string{"cms_key"},
			output: []byte("-ERR wrong number of arguments for 'cms.initbydim' command\r\n"),
		},
		"wrong type of width": {
			input: []string{"cms_key", "not_a_number", "5"},
			output: []byte("-ERR invalid width for 'cms.initbydim' command\r\n"),
		},
		"wrong type of depth": {
			input: []string{"cms_key", "5", "not_a_number"},
			output: []byte("-ERR invalid depth for 'cms.initbydim' command\r\n"),
		},
		"negative width": {
			input: []string{"cms_key", "-100", "5"},
			output: []byte("-ERR invalid width for 'cms.initbydim' command\r\n"),
		},
		"negative depth": {
			input: []string{"cms_key", "5", "-100"},
			output: []byte("-ERR invalid depth for 'cms.initbydim' command\r\n"),
		},
		"correct width and depth": {
			input: []string{"cms_key", "1000", "5"},
			output: clientio.RespOK,
		},
		"key already exists": {
			setup: func () {
				key := "cms_key"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key", "1000", "5"},
			output: []byte("-ERR key already exists for 'cms.initbydim' command\r\n"),
		},
	}

	runEvalTests(t, tests, evalCMSINITBYDIM, store)
}

func testCMSInitByProb(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of arguments": {
			input: []string{"cms_key1"},
			output: []byte("-ERR wrong number of arguments for 'cms.initbyprob' command\r\n"),
		},
		"wrong type of error rate": {
			input: []string{"cms_key1", "not_a_number", "0.01"},
			output: []byte("-ERR invalid overestimation value for 'cms.initbyprob' command\r\n"),
		},
		"wrong type of probability": {
			input: []string{"cms_key1", "0.01", "not_a_number"},
			output: []byte("-ERR invalid prob value for 'cms.initbyprob' command\r\n"),
		},
		"error rate out of range": {
			input: []string{"cms_key1", "1", "0.01"},
			output: []byte("-ERR invalid overestimation value for 'cms.initbyprob' command\r\n"),
		},
		"probability out of range": {
			input: []string{"cms_key1", "0.01", "1"},
			output: []byte("-ERR invalid prob value for 'cms.initbyprob' command\r\n"),
		},
		"correct error rate and probability": {
			input: []string{"cms_key1", "0.01", "0.01"},
			output: clientio.RespOK,
		},
		"key already exists": {
			setup: func () {
				key := "cms_key1"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key1", "0.01", "0.01"},
			output: []byte("-ERR key already exists for 'cms.initbyprob' command\r\n"),
		},
	}

	runEvalTests(t, tests, evalCMSINITBYPROB, store)
}

func testCMSInfo(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of arguments": {
			input: []string{},
			output: []byte("-ERR wrong number of arguments for 'cms.info' command\r\n"),
		},
		"key doesn't exist": {
			input: []string{"cms_key2"},
			output: []byte("-ERR key does not exist for 'cms.info' command\r\n"),
		},
		"one argument": {
			setup: func () {
				key := "cms_key2"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key2"},
			output: clientio.Encode([]interface{}{"width", 1000, "depth", 5, "count", 0}, false),
		},
	}

	runEvalTests(t, tests, evalCMSINFO, store)
}

func testCMSIncrBy(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of arguments": {
			input: []string{"cms_key3"},
			output: []byte("-ERR wrong number of arguments for 'cms.incrby' command\r\n"),
		},
		"key doesn't exist": {
			input: []string{"cms_key3", "test", "10"},
			output: []byte("-ERR key does not exist for 'cms.incrby' command\r\n"),
		},
		"inserting keys": {
			setup: func () {
				key := "cms_key3"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key3", "test", "10", "test1", "10"},
			output: clientio.Encode([]uint64{10,10}, false),
		},
		"missing values": {
			setup: func () {
				key := "cms_key3"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key3", "test", "10", "test1"},
			output: []byte("-ERR wrong number of arguments for 'cms.incrby' command\r\n"),
		},
		"negative values": {
			setup: func () {
				key := "cms_key3"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
			},
			input: []string{"cms_key3", "test", "-1"},
			output: []byte("-ERR cannot parse number for 'cms.incrby' command\r\n"),
		},
	}

	runEvalTests(t, tests, evalCMSIncrBy, store)
}

func testCMSQuery(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of arguments": {
			input: []string{"cms_key4"},
			output: []byte("-ERR wrong number of arguments for 'cms.query' command\r\n"),
		},
		"key doesn't exist": {
			input: []string{"cms_key4", "test"},
			output: []byte("-ERR key does not exist for 'cms.query' command\r\n"),
		},
		"query keys": {
			setup: func () {
				key := "cms_key4"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)
				cms, _ := getCountMinSketch(key, store)

				cms.updateMatrix("test", 10000)
				cms.updateMatrix("test1", 100)
			},
			input: []string{"cms_key4", "test", "test1"},
			output: clientio.Encode([]uint64{10000,100}, false),
		},
	}

	runEvalTests(t, tests, evalCMSQuery, store)
}

func testCMSMerge(t *testing.T, store *dstore.Store) {
	tests := map[string]evalTestCase{
		"wrong number of arguments": {
			input: []string{"cms_key5"},
			output: []byte("-ERR wrong number of arguments for 'cms.merge' command\r\n"),
		},
		"key doesn't exist": {
			input: []string{"cms_key5", "1", "test"},
			output: []byte("-ERR key does not exist for 'cms.merge' command\r\n"),
		},
		"wrong type of number of sources": {
			setup: func () {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

			},
			input: []string{"cms_key5", "not_a_number", "test"},
			output: []byte("-ERR cannot parse number for 'cms.merge' command\r\n"),
		},
		"more sources than specified": {
			setup: func () {
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
			output: []byte("-ERR invalid number of arguments to merge for 'cms.merge' command\r\n"),
		},
		"fewer sources than specified": {
			setup: func () {
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
			output: []byte("-ERR invalid number of arguments to merge for 'cms.merge' command\r\n"),
		},
		"missing weights": {
			setup: func () {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

			},
			input: []string{"cms_key5", "1", "test", "WEIGHTS"},
			output: []byte("-ERR invalid number of arguments to merge for 'cms.merge' command\r\n"),
		},
		"more weights than needed": {
			setup: func () {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

			},
			input: []string{"cms_key5", "1", "test", "WEIGHTS", "1", "2"},
			output: []byte("-ERR invalid number of arguments to merge for 'cms.merge' command\r\n"),
		},
		"correct case": {
			setup: func () {
				key := "cms_key5"
				opts, _ := newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(key, opts, store)

				source1 := "test"
				opts, _ = newCountMinSketchOpts([]string{"1000", "5"})
				createCountMinSketch(source1, opts, store)

			},
			input: []string{"cms_key5", "1", "test"},
			output: clientio.RespOK,
		},
		"correct case with given weights": {
			setup: func () {
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
			output: clientio.RespOK,
		},
	}

	runEvalTests(t, tests, evalCMSMerge, store)
}