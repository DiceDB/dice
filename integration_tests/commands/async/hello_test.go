package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestHello(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	expected := []interface{}{
		"proto", int64(2),
		"id", "0.0.0.0:7379",
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{},
	}

	t.Run("HELLO command response", func(t *testing.T) {
		actual := FireCommand(conn, "HELLO")
		assert.DeepEqual(t, expected, actual)
	})
}
