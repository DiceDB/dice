// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
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

package websocket

import (
	"fmt"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func DeleteKey(t *testing.T, conn *websocket.Conn, exec *WebsocketCommandExecutor, key string) {
	cmd := "DEL " + key
	resp, err := exec.FireCommandAndReadResponse(conn, cmd)
	assert.Nil(t, err)
	respFloat, ok := resp.(float64)
	assert.True(t, ok, "error converting response to float64")
	assert.True(t, respFloat == 1 || respFloat == 0, "unexpected response in %v: %v", cmd, resp)
}

func DeleteHKey(t *testing.T, conn *websocket.Conn, exec *WebsocketCommandExecutor, key, field string) {
	cmd := fmt.Sprintf("HDEL %s %s", key, field)
	resp, err := exec.FireCommandAndReadResponse(conn, cmd)
	assert.Nil(t, err)
	respFloat, ok := resp.(float64)
	assert.True(t, ok, "error converting response to float64")
	assert.True(t, respFloat == 1 || respFloat == 0, "unexpected response in %v: %v", cmd, resp)
}
