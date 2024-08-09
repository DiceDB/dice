package core_test

import (
	"sort"
	"testing"

	"github.com/dicedb/dice/core"
	"github.com/xwb1989/sqlparser"
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
	{"k", "k"},
}

func setup() {
	// delete all keys
	for _, data := range dataset {
		core.Del(data.key)
	}

	for _, data := range dataset {
		core.Put(data.key, &core.Obj{Value: data.value}, nil)
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

	result, err := core.ExecuteQuery(query)

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

	result, err := core.ExecuteQuery(query)

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

	result, err := core.ExecuteQuery(query)

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

	result, err := core.ExecuteQuery(query)

	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(result, 0)) // No keys match "x*"
}

func TestExecuteQueryWithWhere(t *testing.T) {
	setup()
	t.Run("BasicWhereClause", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("v3")),
			},
		}

		result, err := core.ExecuteQuery(query)

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause")
		assert.Equal(t, result[0].Key, "k3")
		assert.DeepEqual(t, result[0].Value.Value, "v3")
	})

	t.Run("EmptyResult", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("nonexistent")),
			},
		}

		result, err := core.ExecuteQuery(query)

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected empty result for non-matching WHERE clause")
	})

	t.Run("ComplexWhereClause", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$value",
				Order:   "desc",
			},
			Where: &sqlparser.AndExpr{
				Left: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: ">",
					Right:    sqlparser.NewStrVal([]byte("v2")),
				},
				Right: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "<",
					Right:    sqlparser.NewStrVal([]byte("v5")),
				},
			},
		}

		result, err := core.ExecuteQuery(query)

		assert.NilError(t, err)
		assert.Equal(t, len(result), 2, "Expected 2 results for complex WHERE clause")
		assert.DeepEqual(t, []string{result[0].Key, result[1].Key}, []string{"k2", "k3"})
	})

	t.Run("ComparingKeyWithValue", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_key")},
				Operator: "=",
				Right:    &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
			},
		}

		result, err := core.ExecuteQuery(query)

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for comparison between key and value")
		assert.Equal(t, result[0].Key, "k")
		assert.DeepEqual(t, result[0].Value.Value, "k")
	})
}

func TestExecuteQueryWithIncompatibleTypes(t *testing.T) {
	setup()

	t.Run("ComparingStrWithInt", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewIntVal([]byte("42")),
			},
		}

		_, err := core.ExecuteQuery(query)

		assert.Error(t, err, "incompatible types in comparison: string and int")
	})

	t.Run("NullValue", func(t *testing.T) {
		// We don't support NULL values in Dice, however, we should include a
		// test for it to ensure the executor handles it correctly.
		core.Put("nullKey", &core.Obj{Value: nil}, nil)
		defer core.Del("nullKey")

		query := core.DSQLQuery{
			KeyRegex: "nullKey",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    &sqlparser.NullVal{},
			},
		}

		_, err := core.ExecuteQuery(query)

		assert.Error(t, err, "unsupported value type: <nil>")
	})
}

func TestExecuteQueryWithEdgeCases(t *testing.T) {
	setup()

	t.Run("CaseSensitivity", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("V3")), // Uppercase V3
			},
		}

		result, err := core.ExecuteQuery(query)

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected 0 results due to case sensitivity")
	})

	t.Run("WhereClauseOnKey", func(t *testing.T) {
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
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_key")},
				Operator: ">",
				Right:    sqlparser.NewStrVal([]byte("k3")),
			},
		}

		result, err := core.ExecuteQuery(query)

		assert.NilError(t, err)
		assert.Equal(t, len(result), 2, "Expected 2 results for WHERE clause on key")
		assert.DeepEqual(t, []string{result[0].Key, result[1].Key}, []string{"k4", "k5"})
	})

	t.Run("UnsupportedOperator", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "LIKE",
				Right:    sqlparser.NewStrVal([]byte("%3")),
			},
		}

		_, err := core.ExecuteQuery(query)

		assert.ErrorContains(t, err, "unsupported operator")
	})

	t.Run("EmptyKeyRegex", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}

		result, err := core.ExecuteQuery(query)

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected no keys to be returned for empty regex")
	})
}
