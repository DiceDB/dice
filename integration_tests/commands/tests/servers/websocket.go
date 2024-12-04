package servers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/querymanager"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/gorilla/websocket"
)

const (
	URL       = "ws://localhost:8380"
	testPort1 = 8380
	testPort2 = 8381
)

type WebsocketCommandExecutor struct {
	baseURL         string
	websocketClient *http.Client
	upgrader        websocket.Upgrader
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

func (e *WebsocketCommandExecutor) FireCommand(conn *websocket.Conn, cmd string) error {
	// send request
	err := conn.WriteMessage(websocket.TextMessage, []byte(cmd))
	if err != nil {
		return err
	}

	return nil
}

func RunWebsocketServer(ctx context.Context, wg *sync.WaitGroup, opt TestServerOptions) {
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Persistence.WriteAOFOnCleanup = false

	// Initialize WebsocketServer
	globalErrChannel := make(chan error)
	watchChan := make(chan dstore.QueryWatchEvent, config.DiceConfig.Performance.WatchChanBufSize)
	shardManager := shard.NewShardManager(1, watchChan, nil, globalErrChannel)
	queryWatcherLocal := querymanager.NewQueryManager()
	config.DiceConfig.WebSocket.Port = opt.Port
	testServer := server.NewWebSocketServer(shardManager, testPort1, nil)
	shardManagerCtx, cancelShardManager := context.WithCancel(ctx)

	// run shard manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(shardManagerCtx)
	}()

	// run query manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		queryWatcherLocal.Run(ctx, watchChan)
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
