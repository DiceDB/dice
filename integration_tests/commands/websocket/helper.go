// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
