package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLPOPCount(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name       string
		commands   []string
		expected   []interface{}
		cleanupKey string
	}{
		{
			name: "LPOP with count argument - valid, invalid, and edge cases",
			commands: []string{
				"RPUSH k v1",
				"RPUSH k v2",
				"RPUSH k v3",
				"RPUSH k v4",
				"LPOP k 2",
				"LPOP k 2",
				"LPOP k -1",
				"LPOP k abc",
				"LLEN k",
			},
			expected: []any{
				float64(1),
				float64(2),
				float64(3),
				float64(4),
				[]interface{}{"v1", "v2"},
				[]interface{}{"v3", "v4"},
				"ERR value is out of range",
				"ERR value is not an integer or out of range",
				float64(0),
			},
			cleanupKey: "k",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()

			for i, cmd := range tc.commands {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
			DeleteKey(t, conn, exec, tc.cleanupKey)
		})
	}
}

