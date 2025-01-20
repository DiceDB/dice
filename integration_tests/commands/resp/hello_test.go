// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package resp

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
		"id", fmt.Sprintf("%s:%d", config.DiceConfig.RespServer.Addr, config.DiceConfig.RespServer.Port),
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{},
	}

	t.Run("HELLO command response", func(t *testing.T) {
		actual := FireCommand(conn, "HELLO")
		assert.Equal(t, expected, actual)
	})
}
