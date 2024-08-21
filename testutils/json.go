package testutils

import (
	"reflect"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/internal/constants"
	"gotest.tools/v3/assert"
)

func IsJSONResponse(s string) bool {
	return (s != constants.EmptyStr && (s[0] == '{' || s[0] == '['))
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

// TODO: Prone to flakiness due to changing order of elements in array. Needs work.
func NormalizeJSON(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		nm := make(map[string]interface{})
		for k, v := range v {
			nm[k] = NormalizeJSON(v)
		}
		return nm
	case []interface{}:
		for i, e := range v {
			v[i] = NormalizeJSON(e)
		}
		return v
	default:
		return v
	}
}
