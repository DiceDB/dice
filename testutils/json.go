package testutils

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/constants"
	"gotest.tools/v3/assert"
)

func IsJSONResponse(s string) bool {
	return (s != constants.EmptyStr && (sonic.ValidString(s)))
}

func AssertJSONEqual(t *testing.T, expected, actual string) {
	var expectedJSON, actualJSON interface{}

	err := sonic.UnmarshalString(expected, &expectedJSON)
	assert.NilError(t, err, "Failed to unmarshal expected JSON")

	err = sonic.UnmarshalString(actual, &actualJSON)
	assert.NilError(t, err, "Failed to unmarshal actual JSON")

	if !reflect.DeepEqual(NormalizeJSON(expectedJSON), NormalizeJSON(actualJSON)) {
		t.Errorf("JSON not equal.\nExpected: %s\nActual: %s", expected, actual)
	}
}

// Rewritten NormalizeJSON to address flakiness by sorting JSON arrays (of comparable types) to ensure consistent ordering.
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
		sort.SliceStable(normalizedArray, func(i, j int) bool {
			return fmt.Sprintf("%v", normalizedArray[i]) < fmt.Sprintf("%v", normalizedArray[j])
		})
		return normalizedArray
	default:
		return v
	}
}
