package sql_test

import (
	"sort"
	"testing"

	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/sql"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/xwb1989/sqlparser"
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
		store.Put(data.key, &object.Obj{Value: data.value})
	}
}

func TestExecuteQueryOrderBykey(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setup(store, simpleKVDataset)

	queryString := "SELECT $key, $value WHERE $key like 'k*' ORDER BY $key ASC"
	query, err := sql.ParseQuery(queryString)
	assert.Nil(t, err)

	result, err := sql.ExecuteQuery(&query, store.GetStore())

	assert.Nil(t, err)
	assert.Equal(t, len(result), len(simpleKVDataset))

	sortedDataset := make([]keyValue, len(simpleKVDataset))
	copy(sortedDataset, simpleKVDataset)

	// Sort the new dataset by the "key" field
	sort.Slice(sortedDataset, func(i, j int) bool {
		return sortedDataset[i].key < sortedDataset[j].key
	})

	for i, data := range sortedDataset {
		assert.Equal(t, result[i].Key, data.key)
		assert.Equal(t, result[i].Value.Value, data.value)
	}
}

func TestExecuteQueryBasicOrderByValue(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setup(store, simpleKVDataset)

	queryStr := "SELECT $key, $value WHERE $key like 'k*' ORDER BY $value ASC"
	query, err := sql.ParseQuery(queryStr)
	assert.Nil(t, err)

	result, err := sql.ExecuteQuery(&query, store.GetStore())

	assert.Nil(t, err)
	assert.Equal(t, len(result), len(simpleKVDataset))

	sortedDataset := make([]keyValue, len(simpleKVDataset))
	copy(sortedDataset, simpleKVDataset)

	// Sort the new dataset by the "value" field
	sort.Slice(sortedDataset, func(i, j int) bool {
		return sortedDataset[i].value < sortedDataset[j].value
	})

	for i, data := range sortedDataset {
		assert.Equal(t, result[i].Key, data.key)
		assert.Equal(t, result[i].Value.Value, data.value)
	}
}

func TestExecuteQueryLimit(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setup(store, simpleKVDataset)

	queryStr := "SELECT $value WHERE $key like 'k*' ORDER BY $key ASC LIMIT 3"
	query, err := sql.ParseQuery(queryStr)
	assert.Nil(t, err)

	result, err := sql.ExecuteQuery(&query, store.GetStore())

	assert.Nil(t, err)
	assert.Equal(t, len(result), 3) // Checks if limit is respected

	sortedDataset := make([]keyValue, len(simpleKVDataset))
	copy(sortedDataset, simpleKVDataset)

	// Sort the new dataset by the "key" field
	sort.Slice(sortedDataset, func(i, j int) bool {
		return sortedDataset[i].key < sortedDataset[j].key
	})

	for i, data := range sortedDataset[:3] {
		assert.Equal(t, result[i].Key, utils.EmptyStr)
		assert.Equal(t, result[i].Value.Value, data.value)
	}
}

func TestExecuteQueryNoMatch(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setup(store, simpleKVDataset)

	queryStr := "SELECT $key, $value WHERE $key like 'x*'"
	query, err := sql.ParseQuery(queryStr)
	assert.Nil(t, err)

	result, err := sql.ExecuteQuery(&query, store.GetStore())

	assert.Nil(t, err)
		assert.Equal(t, len(result), 0) // No keys match "x*"
}

func TestExecuteQueryWithWhere(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setup(store, simpleKVDataset)
	t.Run("BasicWhereClause", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $value = 'v3' AND $key like 'k*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause")
		assert.Equal(t, result[0].Key, "k3")
		assert.Equal(t, result[0].Value.Value, "v3")
	})

	t.Run("EmptyResult", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $value = 'nonexistent' AND $key like 'k*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 0, "Expected empty result for non-matching WHERE clause")
	})

	t.Run("ComplexWhereClause", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $value > 'v2' AND $value < 'v5' AND $key like 'k*' ORDER BY $key ASC"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 2, "Expected 2 results for complex WHERE clause")
		assert.Equal(t, []string{result[0].Key, result[1].Key}, []string{"k2", "k3"})
	})

	t.Run("ComparingKeyWithValue", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $key = $value"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for comparison between key and value")
		assert.Equal(t, result[0].Key, "k")
		assert.Equal(t, result[0].Value.Value, "k")
	})
}

func TestExecuteQueryWithIncompatibleTypes(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setup(store, simpleKVDataset)

	t.Run("ComparingStrWithInt", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $value = 42 AND $key like 'k*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		_, err = sql.ExecuteQuery(&query, store.GetStore())

		assert.Error(t, err, "incompatible types in comparison: string and int64")
	})
}

func TestExecuteQueryWithEdgeCases(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setup(store, simpleKVDataset)

	t.Run("CaseSensitivity", func(t *testing.T) {
		query := sql.DSQLQuery{
			Selection: sql.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("V3")), // Uppercase V3
			},
		}

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 0, "Expected 0 results due to case sensitivity")
	})

	t.Run("WhereClauseOnKey", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $key > 'k3' AND $key like 'k*' ORDER BY $key ASC"
		query, err := sql.ParseQuery(queryStr)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 2, "Expected 2 results for WHERE clause on key")
		assert.Equal(t, []string{result[0].Key, result[1].Key}, []string{"k4", "k5"})
	})

	t.Run("UnsupportedOperator", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $value regexp '%3' AND $key like 'k*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		_, err = sql.ExecuteQuery(&query, store.GetStore())

		assert.ErrorContains(t, err, "unsupported operator")
	})

	t.Run("EmptyKeyRegex", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $key like ''"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
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

		store.Put(data.key, store.NewObj(jsonValue, -1, object.ObjTypeJSON, object.ObjEncodingJSON))
	}
}

func TestExecuteQueryWithJsonExpressionInWhere(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setupJSON(t, store, jsonWhereClauseDataset)

	t.Run("BasicWhereClauseWithJSON", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE '$value.name' = 'Tom' AND $key like 'json*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 results for WHERE clause")
		assert.Equal(t, result[0].Key, "json1")

		var expected, actual interface{}
		assert.Nil(t, sonic.UnmarshalString(`{"name":"Tom"}`, &expected))
		assert.Nil(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.Equal(t, actual, expected)
	})

	t.Run("EmptyResult", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE '$value.name' = 'Bill' AND $key like 'json*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 0, "Expected empty result for non-matching WHERE clause")
	})

	t.Run("WhereClauseWithFloats", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE '$value.score' > 13.15 AND $key like 'json*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with floating point values")
		assert.Equal(t, result[0].Key, "json2")

		var expected, actual interface{}
		assert.Nil(t, sonic.UnmarshalString(`{"name":"Bob","score":18.1}`, &expected))
		assert.Nil(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.Equal(t, actual, expected)
	})

	t.Run("WhereClauseWithInteger", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE '$value.scoreInt' > 13 AND $key like 'json*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with integer values")
		assert.Equal(t, result[0].Key, "json3")

		var expected, actual interface{}
		assert.Nil(t, sonic.UnmarshalString(`{"scoreInt":20}`, &expected))
		assert.Nil(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.Equal(t, actual, expected)
	})

	t.Run("NestedWhereClause", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE '$value.field1.field2.field3.score' < 13 AND $key like 'json*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for WHERE clause with nested json")
		assert.Equal(t, result[0].Key, "json4")

		var expected, actual interface{}
		assert.Nil(t, sonic.UnmarshalString(`{"field1":{"field2":{"field3":{"score":2}}}}`, &expected))
		assert.Nil(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.Equal(t, actual, expected)
	})

	t.Run("ComplexWhereClause", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE '$value.field1.field2.field3.score' > '$value.field1.score2' AND $key like 'json*'"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, len(result), 1, "Expected 1 result for Complex WHERE clause expression")
		assert.Equal(t, result[0].Key, "json5")

		var expected, actual interface{}
		assert.Nil(t, sonic.UnmarshalString(`{"field1":{"field2":{"field3":{"score":18}},"score2":5}}`, &expected))
		assert.Nil(t, sonic.UnmarshalString(result[0].Value.Value.(string), &actual))
		assert.Equal(t, actual, expected)
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
	store := dstore.NewStore(nil, nil)
	setupJSON(t, store, jsonOrderDataset)

	t.Run("OrderBySimpleJSONField", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $key like 'json*' ORDER BY $value.name ASC"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
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
		queryStr := "SELECT $key, $value WHERE $key like 'json*' ORDER BY $value.age DESC"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
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
		queryStr := "SELECT $key, $value WHERE $key like 'json*' ORDER BY '$value.nested.field.value' ASC"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
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
		queryStr := "SELECT $key, $value WHERE $key like 'json*' ORDER BY $value.score DESC"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		// No ordering guarantees for mixed types.
		assert.Equal(t, 5, len(result))
	})

	t.Run("OrderByWithWhereClause", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $key like 'json*' AND '$value.age' > 30 ORDER BY $value.name DESC"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		assert.Equal(t, 3, len(result), "Expected 3 results (age > 30, ordered by name)")
		assert.Equal(t, "json4", result[0].Key) // Eve, age 32
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[4].value, result[0].Value.Value.(string))

		assert.Equal(t, "json5", result[1].Key) // Charlie, age 50
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[3].value, result[1].Value.Value.(string))

		assert.Equal(t, "json3", result[2].Key) // Alice, age 35
		validateJSONStringRepresentationsAreEqual(t, jsonOrderDataset[0].value, result[2].Value.Value.(string))
	})

	t.Run("OrderByNonExistentField", func(t *testing.T) {
		queryStr := "SELECT $key, $value WHERE $key like 'json*' ORDER BY $value.nonexistent ASC"
		query, err := sql.ParseQuery(queryStr)
		assert.Nil(t, err)

		result, err := sql.ExecuteQuery(&query, store.GetStore())

		assert.Nil(t, err)
		// No ordering guarantees for non-existent field references.
		assert.Equal(t, 5, len(result), "Expected 5 results")
	})
}

// validateJSONStringRepresentationsAreEqual unmarshals the expected and actual JSON strings and performs a deep comparison.
func validateJSONStringRepresentationsAreEqual(t *testing.T, expectedJSONString, actualJSONString string) {
	t.Helper()
	var expectedValue, actualValue interface{}
	assert.Nil(t, sonic.UnmarshalString(expectedJSONString, &expectedValue))
	assert.Nil(t, sonic.UnmarshalString(actualJSONString, &actualValue))
	assert.Equal(t, actualValue, expectedValue)
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
	store := dstore.NewStore(nil, nil)
	setup(store, stringComparisonDataset)

	testCases := []struct {
		name       string
		query      string
		expectLen  int
		expectKeys []string
	}{
		{
			name:       "NamesStartingWithA",
			query:      "SELECT $key, $value WHERE $value LIKE 'A*' AND $key LIKE 'user:*'",
			expectLen:  1,
			expectKeys: []string{"user:1"},
		},
		{
			name:       "EmailsWithGmailDomain",
			query:      "SELECT $key, $value WHERE $value LIKE '*@gmail.com' AND $key LIKE 'email:*'",
			expectLen:  1,
			expectKeys: []string{"email:3"},
		},
		{
			name:       "DescriptionsContainingWord",
			query:      "SELECT $key, $value WHERE $value LIKE '*description*' AND $key LIKE 'desc:*' ORDER BY $key ASC",
			expectLen:  2,
			expectKeys: []string{"desc:1", "desc:2"},
		},
		{
			name:       "CaseInsensitiveMatching",
			query:      "SELECT $key, $value WHERE $value LIKE '*UPPERCASE*' AND $key LIKE 'desc:*'",
			expectLen:  1,
			expectKeys: []string{"desc:4"},
		},
		{
			name:       "MatchingSpecialCharacters",
			query:      "SELECT $key, $value WHERE $value LIKE '*!@#*' AND $key LIKE 'desc:*'",
			expectLen:  1,
			expectKeys: []string{"desc:3"},
		},
		{
			name:       "MatchingNumbers",
			query:      "SELECT $key, $value WHERE $value LIKE '*123*' AND $key LIKE 'desc:*'",
			expectLen:  1,
			expectKeys: []string{"desc:3"},
		},
		{
			name:       "ProductsContainingColor",
			query:      "SELECT $key, $value WHERE $value LIKE '*Red*' AND $key LIKE 'product:*'",
			expectLen:  1,
			expectKeys: []string{"product:1"},
		},
		{
			name:       "TagsEndingWithPriority",
			query:      "SELECT $key, $value WHERE $value LIKE '*priority' AND $key LIKE 'tag:*'",
			expectLen:  1,
			expectKeys: []string{"tag:3"},
		},
		{
			name:       "NamesWith5Characters",
			query:      "SELECT $key, $value WHERE $value LIKE '???????????' AND $key LIKE 'user:*' ORDER BY $key ASC",
			expectLen:  2,
			expectKeys: []string{"user:1", "user:2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := sql.ParseQuery(tc.query)
			assert.Nil(t, err)
			result, err := sql.ExecuteQuery(&query, store.GetStore())

			assert.Nil(t, err)
			assert.Equal(t, len(result), tc.expectLen, "Expected %d results, got %d", tc.expectLen, len(result))

			resultKeys := make([]string, len(result))
			for i, r := range result {
				resultKeys[i] = r.Key
			}

			assert.Equal(t, resultKeys, tc.expectKeys)
		})
	}
}

func TestExecuteQueryWithStringNotLikeComparisons(t *testing.T) {
	store := dstore.NewStore(nil, nil)
	setup(store, stringComparisonDataset)

	testCases := []struct {
		name       string
		query      string
		expectLen  int
		expectKeys []string
	}{
		{
			name:       "NamesNotStartingWithA",
			query:      "SELECT $key, $value WHERE $value NOT LIKE 'A*' AND $key LIKE 'user:*' ORDER BY $key ASC",
			expectLen:  4,
			expectKeys: []string{"user:2", "user:3", "user:4", "user:5"},
		},
		{
			name:       "EmailsNotWithGmailDomain",
			query:      "SELECT $key, $value WHERE $value NOT LIKE '*@gmail.com' AND $key LIKE 'email:*' ORDER BY $key ASC",
			expectLen:  4,
			expectKeys: []string{"email:1", "email:2", "email:4", "email:5"},
		},
		{
			name:       "DescriptionsNotContainingWord",
			query:      "SELECT $key, $value WHERE $value NOT LIKE '*description*' AND $key LIKE 'desc:*' ORDER BY $key ASC",
			expectLen:  3,
			expectKeys: []string{"desc:3", "desc:4", "desc:5"},
		},
		{
			name:       "NotCaseInsensitiveMatching",
			query:      "SELECT $key, $value WHERE $value NOT LIKE '*UPPERCASE*' AND $key LIKE 'desc:*' ORDER BY $key ASC",
			expectLen:  4,
			expectKeys: []string{"desc:1", "desc:2", "desc:3", "desc:5"},
		},
		{
			name:       "NotMatchingSpecialCharacters",
			query:      "SELECT $key, $value WHERE $value NOT LIKE '*!@#*' AND $key LIKE 'desc:*' ORDER BY $key ASC",
			expectLen:  4,
			expectKeys: []string{"desc:1", "desc:2", "desc:4", "desc:5"},
		},
		{
			name:       "ProductsNotContainingColor",
			query:      "SELECT $key, $value WHERE $value NOT LIKE '*Red*' AND $key LIKE 'product:*' ORDER BY $key ASC",
			expectLen:  4,
			expectKeys: []string{"product:2", "product:3", "product:4", "product:5"},
		},
		{
			name:       "TagsNotEndingWithPriority",
			query:      "SELECT $key, $value WHERE $value NOT LIKE '*priority' AND $key LIKE 'tag:*' ORDER BY $key ASC",
			expectLen:  4,
			expectKeys: []string{"tag:1", "tag:2", "tag:4", "tag:5"},
		},
		{
			name:       "NamesNotWith5Characters",
			query:      "SELECT $key, $value WHERE $value NOT LIKE '???????????' AND $key LIKE 'user:*' ORDER BY $key ASC",
			expectLen:  3,
			expectKeys: []string{"user:3", "user:4", "user:5"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := sql.ParseQuery(tc.query)
			assert.Nil(t, err)
			result, err := sql.ExecuteQuery(&query, store.GetStore())

			assert.Nil(t, err)
			assert.Equal(t, len(result), tc.expectLen, "Expected %d results, got %d", tc.expectLen, len(result))

			resultKeys := make([]string, len(result))
			for i, r := range result {
				resultKeys[i] = r.Key
			}

			assert.Equal(t, resultKeys, tc.expectKeys)
		})
	}
}
