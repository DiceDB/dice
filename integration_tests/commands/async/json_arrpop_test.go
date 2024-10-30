package async

import (
	"testing"

	"github.com/dicedb/dice/testutils"
	"github.com/stretchr/testify/assert"
)

func TestJSONARRPOP(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	FireCommand(conn, "DEL key")

	arrayAtRoot := `[0,1,2,3]`
	nestedArray := `{"a":2,"b":[0,1,2,3]}`

	testCases := []struct {
		name        string
		commands    []string
		expected    []interface{}
		assertType  []string
		jsonResp    []bool
		nestedArray bool
		path        string
	}{
		{
			name:       "update array at root path",
			commands:   []string{"json.set key $ " + arrayAtRoot, "json.arrpop key $ 2", "json.get key"},
			expected:   []interface{}{"OK", int64(2), "[0,1,3]"},
			assertType: []string{"equal", "equal", "deep_equal"},
		},
		{
			name:       "update nested array",
			commands:   []string{"json.set key $ " + nestedArray, "json.arrpop key $.b 2", "json.get key"},
			expected:   []interface{}{"OK", []interface{}{int64(2)}, `{"a":2,"b":[0,1,3]}`},
			assertType: []string{"equal", "deep_equal", "na"},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := FireCommand(conn, cmd)

				jsonResult, isString := result.(string)

				if isString && testutils.IsJSONResponse(jsonResult) {
					assert.JSONEq(t, out.(string), jsonResult)
					continue
				}

				if tcase.assertType[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assertType[i] == "deep_equal" {
					assert.True(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}
