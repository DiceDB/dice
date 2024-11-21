package async

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/config"
	"github.com/stretchr/testify/assert"
)

func TestHello(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	expected := []interface{}{
		"proto", int64(2),
		"id", fmt.Sprintf("%s:%d", config.DiceConfig.AsyncServer.Addr, config.DiceConfig.AsyncServer.Port),
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{},
	}

	t.Run("HELLO command response", func(t *testing.T) {
		actual := FireCommand(conn, "HELLO")
		assert.Equal(t, expected, actual)
	})
}
