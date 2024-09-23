package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
)

type TestServerOptions struct {
	Port   int
	Logger *slog.Logger
}

type CommandExecutor interface {
	FireCommand(cmd string) interface{}
	Name() string
}

type HTTPCommandExecutor struct {
	httpClient *http.Client
	baseURL    string
}

func NewHTTPCommandExecutor() *HTTPCommandExecutor {
	return &HTTPCommandExecutor{
		baseURL: "http://localhost:8083",
		httpClient: &http.Client{
			Timeout: time.Second * 100,
		},
	}
}

type HTTPCommand struct {
	Command string
	Body    map[string]interface{}
}

func (e *HTTPCommandExecutor) FireCommand(cmd HTTPCommand) interface{} {
	command := strings.ToUpper(cmd.Command)
	body, err := json.Marshal(cmd.Body)

	// Handle error during JSON marshaling
	if err != nil {
		return fmt.Sprintf("ERR failed to marshal command body: %v", err)
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/"+command, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)

	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	var result interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil
	}
	return result
}

func (e *HTTPCommandExecutor) Name() string {
	return "HTTP"
}

func RunHTTPServer(ctx context.Context, wg *sync.WaitGroup, opt TestServerOptions) {
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Server.WriteAOFOnCleanup = false

	watchChan := make(chan dstore.WatchEvent, config.DiceConfig.Server.KeysLimit)
	shardManager := shard.NewShardManager(1, watchChan, opt.Logger)
	config.HTTPPort = opt.Port
	// Initialize the AsyncServer
	testServer := server.NewHTTPServer(shardManager, watchChan, opt.Logger)
	// Inform the user that the server is starting
	fmt.Println("Starting the test server on port", config.HTTPPort)

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
		err := testServer.Run(ctx)
		if err != nil {
			cancelShardManager()
			if errors.Is(err, server.ErrAborted) {
				return
			}
			if err.Error() != "http: Server closed" {
				log.Fatalf("Http test server encountered an error: %v", err)
			}
			log.Printf("Http test server encountered an error: %v", err)
		}
	}()
}
