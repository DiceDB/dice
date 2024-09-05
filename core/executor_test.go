package core_test

import (
	"sort"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/internal/constants"
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

func setup(store *core.Store) {
	// delete all keys
	for _, data := range dataset {
		store.Del(data.key)
	}

	for _, data := range dataset {
		store.Put(data.key, &core.Obj{Value: data.value})
	}
}

func TestExecuteQueryOrderBykey(t *testing.T) {
	store := core.NewStore(nil)
	setup(store)

	query := core.DSQLQuery{
		KeyRegex: "k*",
		Selection: core.QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
		OrderBy: core.QueryOrder{
			OrderBy: "$key",
			Order:   constants.Asc,
		},
	}

	result, err := core.ExecuteQuery(&query, store.GetStore())

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
	store := core.NewStore(nil)
	setup(store)

	query := core.DSQLQuery{
		KeyRegex: "k*",
		Selection: core.QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
		OrderBy: core.QueryOrder{
			OrderBy: "$value",
			Order:   constants.Asc,
		},
	}

	result, err := core.ExecuteQuery(&query, store.GetStore())

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
	store := core.NewStore(nil)
	setup(store)

	query := core.DSQLQuery{
		KeyRegex: "k*",
		Selection: core.QuerySelection{
			KeySelection:   false,
			ValueSelection: true,
		},
		OrderBy: core.QueryOrder{
			OrderBy: "$key",
			Order:   constants.Asc,
		},
		Limit: 3,
	}

	result, err := core.ExecuteQuery(&query, store.GetStore())

	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(result, 3)) // Checks if limit is respected

	sortedDataset := make([]keyValue, len(dataset))
	copy(sortedDataset, dataset)

	// Sort the new dataset by the "key" field
	sort.Slice(sortedDataset, func(i, j int) bool {
		return sortedDataset[i].key < sortedDataset[j].key
	})

	for i, data := range sortedDataset[:3] {
		assert.Equal(t, result[i].Key, constants.EmptyStr)
		assert.DeepEqual(t, result[i].Value.Value, data.value)
	}
}

func TestExecuteQueryNoMatch(t *testing.T) {
	store := core.NewStore(nil)
	setup(store)

	query := core.DSQLQuery{
		KeyRegex: "x*",
		Selection: core.QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
	}

	result, err := core.ExecuteQuery(&query, store.GetStore())

	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(result, 0)) // No keys match "x*"
}

func TestExecuteQueryWithWhere(t *testing.T) {
	store := core.NewStore(nil)
	setup(store)
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

		result, err := core.ExecuteQuery(&query, store.GetStore())

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

		result, err := core.ExecuteQuery(&query, store.GetStore())

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

		result, err := core.ExecuteQuery(&query, store.GetStore())

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

		result, err := core.ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for comparison between key and value")
		assert.Equal(t, result[0].Key, "k")
		assert.DeepEqual(t, result[0].Value.Value, "k")
	})
}

func TestExecuteQueryWithIncompatibleTypes(t *testing.T) {
	store := core.NewStore(nil)
	setup(store)

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

		_, err := core.ExecuteQuery(&query, store.GetStore())

		assert.Error(t, err, "incompatible types in comparison: string and int")
	})

	t.Run("NullValue", func(t *testing.T) {
		// We don't support NULL values in Dice, however, we should include a
		// test for it to ensure the executor handles it correctly.
		store.Put("nullKey", &core.Obj{Value: nil})
		defer store.Del("nullKey")

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

		_, err := core.ExecuteQuery(&query, store.GetStore())

		assert.Error(t, err, "unsupported value type: <nil>")
	})
}

func TestExecuteQueryWithEdgeCases(t *testing.T) {
	store := core.NewStore(nil)
	setup(store)

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

		result, err := core.ExecuteQuery(&query, store.GetStore())

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
				Order:   constants.Asc,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_key")},
				Operator: ">",
				Right:    sqlparser.NewStrVal([]byte("k3")),
			},
		}

		result, err := core.ExecuteQuery(&query, store.GetStore())

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

		_, err := core.ExecuteQuery(&query, store.GetStore())

		assert.ErrorContains(t, err, "unsupported operator")
	})

	t.Run("EmptyKeyRegex", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: constants.EmptyStr,
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}
		result, err := core.ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected no keys to be returned for empty regex")
	})
}

var JsonDataset = []keyValue{
	{"json1", `{"name":"Tom"}`},
	{"json2", `{"name":"Bob","score":18.1}`},
	{"json3", `{"scoreInt":20}`},
	{"json4", `{"field1":{"field2":{"field3":{"score":2}}}}`},
	{"json5", `{"field1":{"field2":{"field3":{"score":18}},"score2":5}}`},
}

func setupJSON(t *testing.T, store *core.Store) {
	t.Helper()
	for _, data := range JsonDataset {
		store.Del(data.key)
	}

	for _, data := range JsonDataset {
		var jsonValue interface{}
		if err := sonic.UnmarshalString(data.value, &jsonValue); err != nil {
			t.Fatalf("Failed to unmarshal value: %v", err)
		}

		store.Put(data.key, store.NewObj(jsonValue, -1, core.ObjTypeJSON, core.ObjEncodingJSON))
	}
}

func TestExecuteQueryWithJsonExpressionInWhere(t *testing.T) {
	store := core.NewStore(nil)
	setupJSON(t, store)

	t.Run("BasicWhereClauseWithJSON", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "json*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.name")),
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("Tom")),
			},
		}

		result, err := core.ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 results for WHERE clause")
		assert.Equal(t, result[0].Key, "json1")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"name":"Tom"}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("EmptyResult", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "json*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.name")),
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("Bill")),
			},
		}

		result, err := core.ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected empty result for non-matching WHERE clause")
	})

	t.Run("WhereClauseWithFloats", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "json*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.score")),
				Operator: ">",
				Right:    sqlparser.NewFloatVal([]byte("13.15")),
			},
		}

		result, err := core.ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with floating point values")
		assert.Equal(t, result[0].Key, "json2")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"name":"Bob","score":18.1}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("WhereClauseWithInteger", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "json*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.scoreInt")),
				Operator: ">",
				Right:    sqlparser.NewIntVal([]byte("13")),
			},
		}

		result, err := core.ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with integer values")
		assert.Equal(t, result[0].Key, "json3")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"scoreInt":20}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("NestedWhereClause", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "json*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.field1.field2.field3.score")),
				Operator: "<",
				Right:    sqlparser.NewIntVal([]byte("13")),
			},
		}

		result, err := core.ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with nested json")
		assert.Equal(t, result[0].Key, "json4")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"field1":{"field2":{"field3":{"score":2}}}}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("ComplexWhereClause", func(t *testing.T) {
		query := core.DSQLQuery{
			KeyRegex: "json*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.field1.field2.field3.score")),
				Operator: ">",
				Right:    sqlparser.NewStrVal([]byte("_value.field1.score2")),
			},
		}

		result, err := core.ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for Complex WHERE clause expression")
		assert.Equal(t, result[0].Key, "json5")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"field1":{"field2":{"field3":{"score":18}},"score2":5}}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})
}
