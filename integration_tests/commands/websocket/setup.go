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

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/gorilla/websocket"
)

const url = "ws://localhost:8380"

type TestServerOptions struct {
	Port   int
	Logger *slog.Logger
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

func NewWebsocketCommandExecutor() *WebsocketCommandExecutor {
	return &WebsocketCommandExecutor{
		baseURL: url,
		websocketClient: &http.Client{
			Timeout: time.Second * 100,
		},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

type WebsocketCommand struct {
	Message map[string]interface{}
}

func (e *WebsocketCommandExecutor) FireCommand(cmd WebsocketCommand) (interface{}, error) {
	msgJSON := cmd.Message

	msg, err := json.Marshal(msgJSON)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}

	// connect with Websocket Server
	conn, res, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with Websocket Server: %v", err)
	}
	defer conn.Close()
	if res != nil {
		res.Body.Close()
	}

	// send request
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return nil, fmt.Errorf("error sending websocket request: %v", err)
	}

	// read the response
	_, resp, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("error reading websocket response: %v", err)
	}

	// marshal to json
	var respJSON interface{}
	if err = json.Unmarshal(resp, &respJSON); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return respJSON, nil
}

func (e *WebsocketCommandExecutor) Name() string {
	return "Websocket"
}

func RunWebsocketServer(ctx context.Context, wg *sync.WaitGroup, opt TestServerOptions) {
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Server.WriteAOFOnCleanup = false

	// Initialize the WebsocketServer
	watchChan := make(chan dstore.WatchEvent, config.DiceConfig.Server.KeysLimit)
	shardManager := shard.NewShardManager(1, watchChan, opt.Logger)
	config.WebsocketPort = opt.Port
	testServer := server.NewWebSocketServer(shardManager, watchChan, opt.Logger)

	shardManagerCtx, cancelShardManager := context.WithCancel(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(shardManagerCtx)
	}()

	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		srverr := testServer.Run(ctx)
		if srverr != nil {
			cancelShardManager()
			if errors.Is(srverr, server.ErrAborted) {
				return
			}
			// if srverr.Error() != "Websocket Server closed" {
			// 	log.Fatalf("Websocket test server encountered an error: %v", srverr)
			// }
			log.Printf("Websocket test server encountered an error: %v", srverr)
		}
	}()
}
