package async

import (
	"reflect"
	"testing"
)

func TestJSONARRINDEX(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	normalArray := `[0,1,2,3,4,3]`
	nestedArray := `{"arrays":[{"arr":[1,2,3]},{"arr":[2,3,4]},{"arr":[1]}]}`

	tests := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "should return error if key is not present",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex nonExistentKey $ 3"},
			expected: []interface{}{"OK", "ERR object with key does not exist"},
		},
		{
			name:     "should return error if json path is invalid",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $invalid_path 3"},
			expected: []interface{}{"OK", "ERR invalid JSONPath"},
		},
		{
			name:     "should return error if provided path does not have any data",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $.some_path 3"},
			expected: []interface{}{"OK", []interface{}{}},
		},
		{
			name:     "should return error if invalid start index provided",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 abc"},
			expected: []interface{}{"OK", "ERR invalid start index"},
		},
		{
			name:     "should return empty list if invalid stop index provided",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 4 abc"},
			expected: []interface{}{"OK", "ERR invalid stop index"},
		},
		{
			name:     "should return array index when given element is present",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $ 3"},
			expected: []interface{}{"OK", []interface{}{int64(3)}},
		},
		{
			name:     "should return -1 when given element is not present",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $ 10"},
			expected: []interface{}{"OK", []interface{}{int64(-1)}},
		},
		{
			name:     "should return array index with start optional param provided",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 4"},
			expected: []interface{}{"OK", []interface{}{int64(5)}},
		},
		{
			name:     "should return array index with start and stop optional param provided",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 4 5"},
			expected: []interface{}{"OK", []interface{}{int64(5)}},
		},
		{
			name:     "should return -1 with start and stop optional param provided where start > stop",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 2 1"},
			expected: []interface{}{"OK", []interface{}{int64(-1)}},
		},
		{
			name:     "should return -1 with start (out of boud) and stop (out of bound) optional param provided",
			commands: []string{"json.set key $ " + normalArray, "json.arrindex key $ 3 6 10"},
			expected: []interface{}{"OK", []interface{}{int64(-1)}},
		},
		{
			name:     "should return list of array indexes for nested json",
			commands: []string{"json.set key $ " + nestedArray, "json.arrindex key $.arrays.*.arr 3"},
			expected: []interface{}{"OK", []interface{}{int64(2), int64(1), int64(-1)}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			FireCommand(conn, "DEL nonExistentKey")
			FireCommand(conn, "DEL key")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				expected := tc.expected[i]

				if !reflect.DeepEqual(result, expected) {
					t.Fatalf("Assertion failed: the expected value is different than result: %s", tc.name)
				}
			}
		})
	}
}
