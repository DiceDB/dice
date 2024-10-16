package websocket

import (
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
