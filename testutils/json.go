package testutils

import (
	"reflect"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/server/utils"
	"gotest.tools/v3/assert"
)

func IsJSONResponse(s string) bool {
	return s != utils.EmptyStr && (sonic.ValidString(s))
}

func AssertJSONEqualList(t *testing.T, expected []string, actual string) {
	res := false
	for _, exp := range expected {
		if compareJSONs(t, exp, actual) {
			res = true
			break
		}
	}
	if !res {
		t.Errorf("JSON not equal.\nExpected one of: %s\nActual: %s", expected, actual)
	}
}

func compareJSONs(t *testing.T, expected, actual string) bool {
	var expectedJSON, actualJSON interface{}

	err := sonic.UnmarshalString(expected, &expectedJSON)
	assert.NilError(t, err, "Failed to unmarshal expected JSON")

	err = sonic.UnmarshalString(actual, &actualJSON)
	assert.NilError(t, err, "Failed to unmarshal actual JSON")

	return reflect.DeepEqual(NormalizeJSON(expectedJSON), NormalizeJSON(actualJSON))
}

func NormalizeJSON(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		normalizedMap := make(map[string]interface{})
		for k, val := range v {
			normalizedMap[k] = NormalizeJSON(val)
		}
		return normalizedMap
	case []interface{}:
		normalizedArray := make([]interface{}, len(v))
		for i, e := range v {
			normalizedArray[i] = NormalizeJSON(e)
		}
		return normalizedArray
	default:
		return v
	}
}
