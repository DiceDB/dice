package tests

import (
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestBgrewriteaof(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	// Defer file removal to ensure cleanup after test completion
	// remove the dice-master.aof file created inside tests folder
	defer os.Remove("dice-master.aof")

	testCases := []TestCase{
		{
			commands: []string{"SET k1 v1", "SET k2 v2", "BGREWRITEAOF"},
			expected: []interface{}{"OK", "OK", "OK"},
		},
	}

	for _, tc := range testCases {
		for i, cmd := range tc.commands {
			result := fireCommand(conn, cmd)
			assert.DeepEqual(t, tc.expected[i], result)
		}

		// ensure that file is written
		time.Sleep(time.Second*2)
		
		fileContent, err := os.ReadFile("dice-master.aof")
		if err != nil {
			t.Fatalf("Failed to read the file: %v", err)
		}

		expectedFileContent, err := os.ReadFile("../testutils/bgrewriteaof-expected-data.aof")
		if err != nil {
			t.Fatalf("Failed to read the file: %v", err)
		}
		
		// Assert that the file content matches the expected data
		assert.Equal(t, string(fileContent), string(expectedFileContent), "File content does not match expected content")
	}
}
