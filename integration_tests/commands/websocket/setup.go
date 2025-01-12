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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/server/httpws"

	"github.com/dicedb/dice/config"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shard"
	"github.com/gorilla/websocket"
)

const (
	URL       = "ws://localhost:8380"
	testPort1 = 8380
	testPort2 = 8381
)

type TestServerOptions struct {
	Port int
}

type CommandExecutor interface {
	FireCommand(cmd string) interface{}
	Name() string
}

type WebsocketCommandExecutor struct {
	baseURL         string
	websocketClient *http.Client
	upgrader        websocket.Upgrader
}

func init() {
	parser := config.NewConfigParser()
	if err := parser.ParseDefaults(config.DiceConfig); err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
}

func NewWebsocketCommandExecutor() *WebsocketCommandExecutor {
	return &WebsocketCommandExecutor{
		baseURL: URL,
		websocketClient: &http.Client{
			Timeout: time.Second * 100,
		},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (e *WebsocketCommandExecutor) ConnectToServer() *websocket.Conn {
	// connect with Websocket Server
	conn, resp, err := websocket.DefaultDialer.Dial(URL, nil)
	if err != nil {
		return nil
	}
	if resp != nil {
		resp.Body.Close()
	}
	return conn
}

func (e *WebsocketCommandExecutor) FireCommandAndReadResponse(conn *websocket.Conn, cmd string) (interface{}, error) {
	err := e.FireCommand(conn, cmd)
	if err != nil {
		return nil, err
	}

	// read the response
	_, resp, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	// marshal to json
	var respJSON interface{}
	if err = json.Unmarshal(resp, &respJSON); err != nil {
		return nil, fmt.Errorf("error unmarshaling response")
	}

	return respJSON, nil
}

func (e *WebsocketCommandExecutor) FireCommand(conn *websocket.Conn, cmd string) error {
	// send request
	err := conn.WriteMessage(websocket.TextMessage, []byte(cmd))
	if err != nil {
		return err
	}

	return nil
}

func (e *WebsocketCommandExecutor) Name() string {
	return "Websocket"
}

func RunWebsocketServer(ctx context.Context, wg *sync.WaitGroup, opt TestServerOptions) {
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Persistence.WriteAOFOnCleanup = false

	// Initialize WebsocketServer
	globalErrChannel := make(chan error)
	shardManager := shard.NewShardManager(1, nil, globalErrChannel)
	config.DiceConfig.WebSocket.Port = opt.Port
	testServer := httpws.NewWebSocketServer(shardManager, testPort1, nil)
	shardManagerCtx, cancelShardManager := context.WithCancel(ctx)

	// run shard manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(shardManagerCtx)
	}()

	// start websocket server
	wg.Add(1)
	go func() {
		defer wg.Done()
		srverr := testServer.Run(ctx)
		if srverr != nil {
			cancelShardManager()
			if errors.Is(srverr, derrors.ErrAborted) {
				return
			}
			slog.Debug("Websocket test server encountered an error: %v", slog.Any("error", srverr))
		}
	}()
}
