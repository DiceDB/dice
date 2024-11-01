package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONRESP(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()
	FireCommand(conn, "DEL key")

	arrayAtRoot := `["dice",10,10.5,true,null]`
	object := `{"b":["dice",10,10.5,true,null]}`

	testCases := []struct {
		name        string
		commands    []string
		expected    []interface{}
		assert_type []string
		jsonResp    []bool
		nestedArray bool
		path        string
	}{
		{
			name:        "print array with mixed types",
			commands:    []string{"json.set key $ " + arrayAtRoot, "json.resp key $"},
			expected:    []interface{}{"OK", []interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"}},
			assert_type: []string{"equal", "equal"},
		},
		{
			name:        "print nested array with mixed types",
			commands:    []string{"json.set key $ " + object, "json.resp key $.b"},
			expected:    []interface{}{"OK", []interface{}{[]interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"}}},
			assert_type: []string{"equal", "equal"},
		},
		{
			name:        "print object at root path",
			commands:    []string{"json.set key $ " + object, "json.resp key"},
			expected:    []interface{}{"OK", []interface{}{"{", "b", []interface{}{"[", "dice", int64(10), "10.5", "true", "(nil)"}}},
			assert_type: []string{"equal", "equal"},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			for i := 0; i < len(tcase.commands); i++ {
				cmd := tcase.commands[i]
				out := tcase.expected[i]
				result := FireCommand(conn, cmd)

				if tcase.assert_type[i] == "equal" {
					assert.Equal(t, out, result)
				} else if tcase.assert_type[i] == "deep_equal" {
					assert.True(t, arraysArePermutations(out.([]interface{}), result.([]interface{})))
				}
			}
		})
	}
}
