package async

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestSelect(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	t.Run("SELECT command response", func(t *testing.T) {
		actual := FireCommand(conn, "SELECT 1")
		assert.DeepEqual(t, "OK", actual)
	})

	t.Run("SELECT command error response", func(t *testing.T) {
		actual := FireCommand(conn, "SELECT")
		assert.DeepEqual(t, "ERR wrong number of arguments for 'select' command", actual)
	})
}
