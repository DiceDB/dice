// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
