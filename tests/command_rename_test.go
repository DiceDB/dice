package tests

import (
	"testing"

	"gotest.tools/v3/assert"
)

var renameKeysTestCases = []struct {
	name     string
	inCmd    []string
	expected []interface{}
}{
	{"Set key and Rename key", []string{"set sourceKey hello", "get sourceKey", "rename sourceKey destKey", "get destKey", "get sourceKey"}, []interface{}{"OK", "hello", "OK", "hello", "(nil)"}},
	{"same key for source and destination on Rename", []string{"set Key hello", "get Key", "rename Key Key", "get Key"}, []interface{}{"OK", "hello", "OK", "hello"}},
	{"If source key doesn't exists", []string{"rename unknownKey Key"}, []interface{}{"(error) ERR no such key"}},
	{"If destination Key already presents", []string{"set destinationKey world", "set newKey hello", "rename newKey destinationKey", "get newKey", "get destinationKey"}, []interface{}{"OK", "OK", "OK", "(nil)", "hello"}},
}

func TestCommandRename(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	for _, tc := range renameKeysTestCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k", "k1", "k2"})
			for i, cmd := range tc.inCmd {
				result := fireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
