package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var renameKeysTestCases = []struct {
	name     string
	inCmd    []string
	expected []interface{}
}{
	{
		name:     "Set key and Rename key",
		inCmd:    []string{"set sourceKey hello", "get sourceKey", "rename sourceKey destKey", "get destKey", "get sourceKey"},
		expected: []interface{}{"OK", "hello", "OK", "hello", "(nil)"},
	},
	{
		name:     "same key for source and destination on Rename",
		inCmd:    []string{"set Key hello", "get Key", "rename Key Key", "get Key"},
		expected: []interface{}{"OK", "hello", "OK", "hello"},
	},
	{
		name:     "If source key doesn't exists",
		inCmd:    []string{"rename unknownKey Key"},
		expected: []interface{}{"ERR no such key"},
	},
	{
		name:     "If source key doesn't exists and renaming the same key to the same key",
		inCmd:    []string{"rename unknownKey unknownKey"},
		expected: []interface{}{"ERR no such key"},
	},
	{
		name:     "If destination Key already presents",
		inCmd:    []string{"set destinationKey world", "set newKey hello", "rename newKey destinationKey", "get newKey", "get destinationKey"},
		expected: []interface{}{"OK", "OK", "OK", "(nil)", "hello"},
	},
}

func TestCommandRename(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range renameKeysTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// deleteTestKeys([]string{"k", "k1", "k2"}, store)
			FireCommand(conn, "DEL k1")
			FireCommand(conn, "DEL k2")
			FireCommand(conn, "DEL 3")
			for i, cmd := range tc.inCmd {
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
