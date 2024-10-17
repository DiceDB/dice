package resp

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHexists(t *testing.T){
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []TestCase{

	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"k"}, store)
			FireCommand(conn, "DEL k")

			for i, cmd := range tc.commands {
				result := FireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}