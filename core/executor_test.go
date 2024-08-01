package core_test

import (
	"sort"
	"testing"

	"github.com/dicedb/dice/core"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type keyValue struct {
	key   string
	value string
}

var dataset = []keyValue{
	{"k2", "v4"},
	{"k4", "v2"},
	{"k3", "v3"},
	{"k5", "v1"},
	{"k1", "v5"},
}

var store = core.NewStore()

func setup() {
	// delete all keys
	for _, data := range dataset {
		store.Del(data.key)
	}

	for _, data := range dataset {
		store.Put(data.key, &core.Obj{Value: data.value}, -1)
	}
}

func TestExecuteQueryOrderBykey(t *testing.T) {
	setup()

	query := core.DSQLQuery{
		KeyRegex: "k*",
		Selection: core.QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
		OrderBy: core.QueryOrder{
			OrderBy: "$key",
			Order:   "asc",
		},
	}

	result, err := core.ExecuteQuery(store, query)

	assert.NilError(t, err)
	assert.Equal(t, len(result), len(dataset))

	sortedDataset := make([]keyValue, len(dataset))
	copy(sortedDataset, dataset)

	// Sort the new dataset by the "key" field
	sort.Slice(sortedDataset, func(i, j int) bool {
		return sortedDataset[i].key < sortedDataset[j].key
	})

	for i, data := range sortedDataset {
		assert.Equal(t, result[i].Key, data.key)
		assert.DeepEqual(t, result[i].Value.Value, data.value)
	}
}

func TestExecuteQueryBasicOrderByValue(t *testing.T) {
	setup()

	query := core.DSQLQuery{
		KeyRegex: "k*",
		Selection: core.QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
		OrderBy: core.QueryOrder{
			OrderBy: "$value",
			Order:   "asc",
		},
	}

	result, err := core.ExecuteQuery(store, query)

	assert.NilError(t, err)
	assert.Equal(t, len(result), len(dataset))

	sortedDataset := make([]keyValue, len(dataset))
	copy(sortedDataset, dataset)

	// Sort the new dataset by the "value" field
	sort.Slice(sortedDataset, func(i, j int) bool {
		return sortedDataset[i].value < sortedDataset[j].value
	})

	for i, data := range sortedDataset {
		assert.Equal(t, result[i].Key, data.key)
		assert.DeepEqual(t, result[i].Value.Value, data.value)
	}
}

func TestExecuteQueryLimit(t *testing.T) {
	setup()

	query := core.DSQLQuery{
		KeyRegex: "k*",
		Selection: core.QuerySelection{
			KeySelection:   false,
			ValueSelection: true,
		},
		OrderBy: core.QueryOrder{
			OrderBy: "$key",
			Order:   "asc",
		},
		Limit: 3,
	}

	result, err := core.ExecuteQuery(store, query)

	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(result, 3)) // Checks if limit is respected

	sortedDataset := make([]keyValue, len(dataset))
	copy(sortedDataset, dataset)

	// Sort the new dataset by the "key" field
	sort.Slice(sortedDataset, func(i, j int) bool {
		return sortedDataset[i].key < sortedDataset[j].key
	})

	for i, data := range sortedDataset[:3] {
		assert.Equal(t, result[i].Key, "")
		assert.DeepEqual(t, result[i].Value.Value, data.value)
	}
}

func TestExecuteQueryNoMatch(t *testing.T) {
	setup()

	query := core.DSQLQuery{
		KeyRegex: "x*",
		Selection: core.QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
	}

	result, err := core.ExecuteQuery(store, query)

	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(result, 0)) // No keys match "x*"
}
