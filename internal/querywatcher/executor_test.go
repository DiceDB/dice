package querywatcher

import (
	"sort"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/xwb1989/sqlparser"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type keyValue struct {
	key   string
	value string
}

var (
	simpleKVDataset = []keyValue{
		{"k2", "v4"},
		{"k4", "v2"},
		{"k3", "v3"},
		{"k5", "v1"},
		{"k1", "v5"},
		{"k", "k"},
	}
)

func setup(store *dstore.Store, dataset []keyValue) {
	// delete all keys
	for _, data := range dataset {
		store.Del(data.key)
	}

	for _, data := range dataset {
		store.Put(data.key, &dstore.Obj{Value: data.value})
	}
}

func TestExecuteQueryOrderBykey(t *testing.T) {
	store := dstore.NewStore(nil)
	setup(store, simpleKVDataset)

	query := DSQLQuery{
		KeyRegex: "k*",
		Selection: QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
		OrderBy: QueryOrder{
			OrderBy: "_key",
			Order:   Asc,
		},
	}

	result, err := ExecuteQuery(&query, store.GetStore())

	assert.NilError(t, err)
	assert.Equal(t, len(result), len(simpleKVDataset))

	sortedDataset := make([]keyValue, len(simpleKVDataset))
	copy(sortedDataset, simpleKVDataset)

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
	store := dstore.NewStore(nil)
	setup(store, simpleKVDataset)

	query := DSQLQuery{
		KeyRegex: "k*",
		Selection: QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
		OrderBy: QueryOrder{
			OrderBy: "_value",
			Order:   Asc,
		},
	}

	result, err := ExecuteQuery(&query, store.GetStore())

	assert.NilError(t, err)
	assert.Equal(t, len(result), len(simpleKVDataset))

	sortedDataset := make([]keyValue, len(simpleKVDataset))
	copy(sortedDataset, simpleKVDataset)

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
	store := dstore.NewStore(nil)
	setup(store, simpleKVDataset)

	query := DSQLQuery{
		KeyRegex: "k*",
		Selection: QuerySelection{
			KeySelection:   false,
			ValueSelection: true,
		},
		OrderBy: QueryOrder{
			OrderBy: "_key",
			Order:   Asc,
		},
		Limit: 3,
	}

	result, err := ExecuteQuery(&query, store.GetStore())

	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(result, 3)) // Checks if limit is respected

	sortedDataset := make([]keyValue, len(simpleKVDataset))
	copy(sortedDataset, simpleKVDataset)

	// Sort the new dataset by the "key" field
	sort.Slice(sortedDataset, func(i, j int) bool {
		return sortedDataset[i].key < sortedDataset[j].key
	})

	for i, data := range sortedDataset[:3] {
		assert.Equal(t, result[i].Key, utils.EmptyStr)
		assert.DeepEqual(t, result[i].Value.Value, data.value)
	}
}

func TestExecuteQueryNoMatch(t *testing.T) {
	store := dstore.NewStore(nil)
	setup(store, simpleKVDataset)

	query := DSQLQuery{
		KeyRegex: "x*",
		Selection: QuerySelection{
			KeySelection:   true,
			ValueSelection: true,
		},
	}

	result, err := ExecuteQuery(&query, store.GetStore())

	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(result, 0)) // No keys match "x*"
}

func TestExecuteQueryWithWhere(t *testing.T) {
	store := dstore.NewStore(nil)
	setup(store, simpleKVDataset)
	t.Run("BasicWhereClause", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("v3")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause")
		assert.Equal(t, result[0].Key, "k3")
		assert.DeepEqual(t, result[0].Value.Value, "v3")
	})

	t.Run("EmptyResult", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("nonexistent")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected empty result for non-matching WHERE clause")
	})

	t.Run("ComplexWhereClause", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "_value",
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

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 2, "Expected 2 results for complex WHERE clause")
		assert.DeepEqual(t, []string{result[0].Key, result[1].Key}, []string{"k2", "k3"})
	})

	t.Run("ComparingKeyWithValue", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_key")},
				Operator: "=",
				Right:    &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for comparison between key and value")
		assert.Equal(t, result[0].Key, "k")
		assert.DeepEqual(t, result[0].Value.Value, "k")
	})
}

func TestExecuteQueryWithIncompatibleTypes(t *testing.T) {
	store := dstore.NewStore(nil)
	setup(store, simpleKVDataset)

	t.Run("ComparingStrWithInt", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewIntVal([]byte("42")),
			},
		}

		_, err := ExecuteQuery(&query, store.GetStore())

		assert.Error(t, err, "incompatible types in comparison: string and int64")
	})

	t.Run("NullValue", func(t *testing.T) {
		// We don't support NULL values in Dice, however, we should include a
		// test for it to ensure the executor handles it correctly.
		store.Put("nullKey", &dstore.Obj{Value: nil})
		defer store.Del("nullKey")

		query := DSQLQuery{
			KeyRegex: "nullKey",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    &sqlparser.NullVal{},
			},
		}

		_, err := ExecuteQuery(&query, store.GetStore())

		assert.Error(t, err, "unsupported value type: <nil>")
	})
}

func TestExecuteQueryWithEdgeCases(t *testing.T) {
	store := dstore.NewStore(nil)
	setup(store, simpleKVDataset)

	t.Run("CaseSensitivity", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("V3")), // Uppercase V3
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected 0 results due to case sensitivity")
	})

	t.Run("WhereClauseOnKey", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "_key",
				Order:   Asc,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_key")},
				Operator: ">",
				Right:    sqlparser.NewStrVal([]byte("k3")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 2, "Expected 2 results for WHERE clause on key")
		assert.DeepEqual(t, []string{result[0].Key, result[1].Key}, []string{"k4", "k5"})
	})

	t.Run("UnsupportedOperator", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "LIKE",
				Right:    sqlparser.NewStrVal([]byte("%3")),
			},
		}

		_, err := ExecuteQuery(&query, store.GetStore())

		assert.ErrorContains(t, err, "unsupported operator")
	})

	t.Run("EmptyKeyRegex", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: utils.EmptyStr,
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}
		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected no keys to be returned for empty regex")
	})
}

var jsonWhereClauseDataset = []keyValue{
	{"json1", `{"name":"Tom"}`},
	{"json2", `{"name":"Bob","score":18.1}`},
	{"json3", `{"scoreInt":20}`},
	{"json4", `{"field1":{"field2":{"field3":{"score":2}}}}`},
	{"json5", `{"field1":{"field2":{"field3":{"score":18}},"score2":5}}`},
}

func setupJSON(t *testing.T, store *dstore.Store, dataset []keyValue) {
	t.Helper()
	for _, data := range dataset {
		store.Del(data.key)
	}

	for _, data := range dataset {
		var jsonValue interface{}
		if err := sonic.UnmarshalString(data.value, &jsonValue); err != nil {
			t.Fatalf("Failed to unmarshal value: %v", err)
		}

		store.Put(data.key, store.NewObj(jsonValue, -1, dstore.ObjTypeJSON, dstore.ObjEncodingJSON))
	}
}

func TestExecuteQueryWithJsonExpressionInWhere(t *testing.T) {
	store := dstore.NewStore(nil)
	setupJSON(t, store, jsonWhereClauseDataset)

	t.Run("BasicWhereClauseWithJSON", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.name")),
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("Tom")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 results for WHERE clause")
		assert.Equal(t, result[0].Key, "json1")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"name":"Tom"}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("EmptyResult", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.name")),
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("Bill")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 0, "Expected empty result for non-matching WHERE clause")
	})

	t.Run("WhereClauseWithFloats", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.score")),
				Operator: ">",
				Right:    sqlparser.NewFloatVal([]byte("13.15")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with floating point values")
		assert.Equal(t, result[0].Key, "json2")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"name":"Bob","score":18.1}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("WhereClauseWithInteger", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.scoreInt")),
				Operator: ">",
				Right:    sqlparser.NewIntVal([]byte("13")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with integer values")
		assert.Equal(t, result[0].Key, "json3")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"scoreInt":20}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("NestedWhereClause", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.field1.field2.field3.score")),
				Operator: "<",
				Right:    sqlparser.NewIntVal([]byte("13")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with nested json")
		assert.Equal(t, result[0].Key, "json4")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"field1":{"field2":{"field3":{"score":2}}}}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})

	t.Run("ComplexWhereClause", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.field1.field2.field3.score")),
				Operator: ">",
				Right:    sqlparser.NewStrVal([]byte("_value.field1.score2")),
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for Complex WHERE clause expression")
		assert.Equal(t, result[0].Key, "json5")

		var expected, actual interface{}
		assert.NilError(t, sonic.UnmarshalString(`{"field1":{"field2":{"field3":{"score":18}},"score2":5}}`, &expected))
		assert.NilError(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.DeepEqual(t, actual, expected)
	})
}

var jsonOrderDataset = []keyValue{
	{"json3", `{"name":"Alice", "age":35, "scoreInt":20, "nested":{"field":{"value":15}}}`},
	{"json2", `{"name":"Bob", "age":25, "score":18.1, "nested":{"field":{"value":40}}}`},
	{"json1", `{"name":"Tom", "age":30, "nested":{"field":{"value":20}}}`},
	{"json5", `{"name":"Charlie", "age":50, "nested":{"field":{"value":19}}}`},
	{"json4", `{"name":"Eve", "age":32, "nested":{"field":{"value":60}}}`},
}

func TestExecuteQueryWithJsonOrderBy(t *testing.T) {
	store := dstore.NewStore(nil)
	setupJSON(t, store, jsonOrderDataset)

	t.Run("OrderBySimpleJSONField", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "_value.name",
				Order:   "asc",
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, 5, len(result), "Expected 5 results")

		assert.Equal(t, "json3", result[0].Key) // Alice
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[0].value, result[0].Value.Value.(string))

		assert.Equal(t, "json2", result[1].Key) // Bob
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[1].value, result[1].Value.Value.(string))

		assert.Equal(t, "json5", result[2].Key) // Charlie
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[3].value, result[2].Value.Value.(string))

		assert.Equal(t, "json4", result[3].Key) // Eve
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[4].value, result[3].Value.Value.(string))

		assert.Equal(t, "json1", result[4].Key) // Tom
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[2].value, result[4].Value.Value.(string))
	})

	t.Run("OrderByNumericJSONField", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "_value.age",
				Order:   "desc",
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, 5, len(result))

		assert.Equal(t, "json5", result[0].Key) // Charlie, age 50
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[3].value, result[0].Value.Value.(string))

		assert.Equal(t, "json3", result[1].Key) // Alice, age 35
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[0].value, result[1].Value.Value.(string))

		assert.Equal(t, "json4", result[2].Key) // Eve, age 32
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[4].value, result[2].Value.Value.(string))

		assert.Equal(t, "json1", result[3].Key) // Tom, age 30
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[2].value, result[3].Value.Value.(string))

		assert.Equal(t, "json2", result[4].Key) // Bob, age 25
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[1].value, result[4].Value.Value.(string))
	})

	t.Run("OrderByNestedJSONField", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "_value.nested.field.value",
				Order:   "asc",
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, 5, len(result))
		assert.Equal(t, "json3", result[0].Key) // Alice, nested.field.value: 15
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[0].value, result[0].Value.Value.(string))

		assert.Equal(t, "json5", result[1].Key) // Charlie, nested.field.value: 19
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[3].value, result[1].Value.Value.(string))

		assert.Equal(t, "json1", result[2].Key) // Tom, nested.field.value: 20
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[2].value, result[2].Value.Value.(string))

		assert.Equal(t, "json2", result[3].Key) // Bob, nested.field.value: 40
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[1].value, result[3].Value.Value.(string))

		assert.Equal(t, "json4", result[4].Key) // Eve, nested.field.value: 60
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[4].value, result[4].Value.Value.(string))
	})

	t.Run("OrderByMixedTypes", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "_value.score",
				Order:   "desc",
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		// No ordering guarantees for mixed types.
		assert.Equal(t, 5, len(result))
	})

	t.Run("OrderByWithWhereClause", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     sqlparser.NewStrVal([]byte("_value.age")),
				Operator: ">",
				Right:    sqlparser.NewIntVal([]byte("30")),
			},
			OrderBy: QueryOrder{
				OrderBy: "_value.name",
				Order:   "desc",
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		assert.Equal(t, 3, len(result), "Expected 3 results (age > 30, ordered by name)")
		assert.Equal(t, "json4", result[0].Key) // Eve, age 32
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[4].value, result[0].Value.Value.(string))

		assert.Equal(t, "json5", result[1].Key) // Charlie, age 50
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[3].value, result[1].Value.Value.(string))

		assert.Equal(t, "json3", result[2].Key) // Alice, age 35
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[0].value, result[2].Value.Value.(string))
	})

	t.Run("OrderByNonExistentField", func(t *testing.T) {
		query := DSQLQuery{
			KeyRegex: "json*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "_value.nonexistent",
				Order:   "asc",
			},
		}

		result, err := ExecuteQuery(&query, store.GetStore())

		assert.NilError(t, err)
		// No ordering guarantees for non-existent field references.
		assert.Equal(t, 5, len(result), "Expected 5 results")
	})
}

// validateJSONStringRepresentationsAreEqual unmarshals the expected and actual JSON strings and performs a deep comparison.
func validateJSONStringRepresentationsAreEqual(t *testing.T, expectedJSONString, actualJSONString string) {
	t.Helper()
	var expectedValue, actualValue interface{}
	assert.NilError(t, sonic.UnmarshalString(expectedJSONString, &expectedValue))
	assert.NilError(t, sonic.UnmarshalString(actualJSONString, &actualValue))
	assert.DeepEqual(t, actualValue, expectedValue)
}

// Dataset will be used for LIKE comparisons
var stringComparisonDataset = []keyValue{
	{"user:1", "Alice Smith"},
	{"user:2", "Bob Johnson"},
	{"user:3", "Charlie Brown"},
	{"user:4", "David Lee"},
	{"user:5", "Eve Wilson"},
	{"product:1", "Red Apple"},
	{"product:2", "Green Banana"},
	{"product:3", "Yellow Lemon"},
	{"product:4", "Orange Orange"},
	{"product:5", "Purple Grape"},
	{"email:1", "alice@example.com"},
	{"email:2", "bob@test.org"},
	{"email:3", "charlie@gmail.com"},
	{"email:4", "david@company.net"},
	{"email:5", "eve@domain.io"},
	{"desc:1", "This is a short description"},
	{"desc:2", "A slightly longer description with more words"},
	{"desc:3", "Description containing numbers 123 and symbols !@#"},
	{"desc:4", "UPPERCASE DESCRIPTION"},
	{"desc:5", "mixed CASE DeScRiPtIoN"},
	{"tag:1", "important"},
	{"tag:2", "urgent"},
	{"tag:3", "low-priority"},
	{"tag:4", "follow-up"},
	{"tag:5", "archived"},
}

func TestExecuteQueryWithLikeStringComparisons(t *testing.T) {
	store := dstore.NewStore(nil)
	setup(store, stringComparisonDataset)

	testCases := []struct {
		name       string
		query      DSQLQuery
		expectLen  int
		expectKeys []string
	}{
		{
			name: "NamesStartingWithA",
			query: DSQLQuery{
				KeyRegex: "user:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("A*")),
				},
			},
			expectLen:  1,
			expectKeys: []string{"user:1"},
		},
		{
			name: "EmailsWithGmailDomain",
			query: DSQLQuery{
				KeyRegex: "email:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("*@gmail.com")),
				},
			},
			expectLen:  1,
			expectKeys: []string{"email:3"},
		},
		{
			name: "DescriptionsContainingWord",
			query: DSQLQuery{
				KeyRegex: "desc:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("*description*")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  2,
			expectKeys: []string{"desc:1", "desc:2"},
		},
		{
			name: "CaseInsensitiveMatching",
			query: DSQLQuery{
				KeyRegex: "desc:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("*UPPERCASE*")),
				},
			},
			expectLen:  1,
			expectKeys: []string{"desc:4"},
		},
		{
			name: "MatchingSpecialCharacters",
			query: DSQLQuery{
				KeyRegex: "desc:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("*!@#*")),
				},
			},
			expectLen:  1,
			expectKeys: []string{"desc:3"},
		},
		{
			name: "MatchingNumbers",
			query: DSQLQuery{
				KeyRegex: "desc:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("*123*")),
				},
			},
			expectLen:  1,
			expectKeys: []string{"desc:3"},
		},
		{
			name: "ProductsContainingColor",
			query: DSQLQuery{
				KeyRegex: "product:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("*Red*")),
				},
			},
			expectLen:  1,
			expectKeys: []string{"product:1"},
		},
		{
			name: "TagsEndingWithPriority",
			query: DSQLQuery{
				KeyRegex: "tag:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("*priority")),
				},
			},
			expectLen:  1,
			expectKeys: []string{"tag:3"},
		},
		{
			name: "NamesWith5Characters",
			query: DSQLQuery{
				KeyRegex: "user:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "LIKE",
					Right:    sqlparser.NewStrVal([]byte("???????????")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  2,
			expectKeys: []string{"user:1", "user:2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ExecuteQuery(&tc.query, store.GetStore())

			assert.NilError(t, err)
			assert.Equal(t, len(result), tc.expectLen, "Expected %d results, got %d", tc.expectLen, len(result))

			resultKeys := make([]string, len(result))
			for i, r := range result {
				resultKeys[i] = r.Key
			}

			assert.DeepEqual(t, resultKeys, tc.expectKeys)
		})
	}
}

func TestExecuteQueryWithStringNotLikeComparisons(t *testing.T) {
	store := dstore.NewStore(nil)
	setup(store, stringComparisonDataset)

	testCases := []struct {
		name       string
		query      DSQLQuery
		expectLen  int
		expectKeys []string
	}{
		{
			name: "NamesNotStartingWithA",
			query: DSQLQuery{
				KeyRegex: "user:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "NOT LIKE",
					Right:    sqlparser.NewStrVal([]byte("A*")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  4,
			expectKeys: []string{"user:2", "user:3", "user:4", "user:5"},
		},
		{
			name: "EmailsNotWithGmailDomain",
			query: DSQLQuery{
				KeyRegex: "email:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "NOT LIKE",
					Right:    sqlparser.NewStrVal([]byte("*@gmail.com")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  4,
			expectKeys: []string{"email:1", "email:2", "email:4", "email:5"},
		},
		{
			name: "DescriptionsNotContainingWord",
			query: DSQLQuery{
				KeyRegex: "desc:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "NOT LIKE",
					Right:    sqlparser.NewStrVal([]byte("*description*")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  3,
			expectKeys: []string{"desc:3", "desc:4", "desc:5"},
		},
		{
			name: "NotCaseInsensitiveMatching",
			query: DSQLQuery{
				KeyRegex: "desc:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "NOT LIKE",
					Right:    sqlparser.NewStrVal([]byte("*UPPERCASE*")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  4,
			expectKeys: []string{"desc:1", "desc:2", "desc:3", "desc:5"},
		},
		{
			name: "NotMatchingSpecialCharacters",
			query: DSQLQuery{
				KeyRegex: "desc:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "NOT LIKE",
					Right:    sqlparser.NewStrVal([]byte("*!@#*")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  4,
			expectKeys: []string{"desc:1", "desc:2", "desc:4", "desc:5"},
		},
		{
			name: "ProductsNotContainingColor",
			query: DSQLQuery{
				KeyRegex: "product:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "NOT LIKE",
					Right:    sqlparser.NewStrVal([]byte("*Red*")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  4,
			expectKeys: []string{"product:2", "product:3", "product:4", "product:5"},
		},
		{
			name: "TagsNotEndingWithPriority",
			query: DSQLQuery{
				KeyRegex: "tag:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "NOT LIKE",
					Right:    sqlparser.NewStrVal([]byte("*priority")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  4,
			expectKeys: []string{"tag:1", "tag:2", "tag:4", "tag:5"},
		},
		{
			name: "NamesNotWith5Characters",
			query: DSQLQuery{
				KeyRegex: "user:*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "NOT LIKE",
					Right:    sqlparser.NewStrVal([]byte("???????????")),
				},
				OrderBy: QueryOrder{
					OrderBy: "_key",
					Order:   Asc,
				},
			},
			expectLen:  3,
			expectKeys: []string{"user:3", "user:4", "user:5"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ExecuteQuery(&tc.query, store.GetStore())

			assert.NilError(t, err)
			assert.Equal(t, len(result), tc.expectLen, "Expected %d results, got %d", tc.expectLen, len(result))

			resultKeys := make([]string, len(result))
			for i, r := range result {
				resultKeys[i] = r.Key
			}

			assert.DeepEqual(t, resultKeys, tc.expectKeys)
		})
	}
}
