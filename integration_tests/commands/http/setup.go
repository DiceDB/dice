package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/server/utils"

	"github.com/dicedb/dice/config"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/querymanager"
	"github.com/dicedb/dice/internal/server"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
)

type TestServerOptions struct {
	Port int
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

func (e *HTTPCommandExecutor) FireCommand(cmd HTTPCommand) (interface{}, error) {
	command := strings.ToUpper(cmd.Command)
	var body []byte
	if cmd.Body != nil {
		var err error
		body, err = json.Marshal(cmd.Body)
		// Handle error during JSON marshaling
		if err != nil {
			return nil, err
		}
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/"+command, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if cmd.Command != "Q.WATCH" {
		var result utils.HTTPResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, err
		}

		return result.Data, nil
	}
	var result interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (e *HTTPCommandExecutor) Name() string {
	return "HTTP"
}

func RunHTTPServer(ctx context.Context, wg *sync.WaitGroup, opt TestServerOptions) {
	config.DiceConfig.Network.IOBufferLength = 16
	config.DiceConfig.Persistence.WriteAOFOnCleanup = false

	globalErrChannel := make(chan error)
	watchChan := make(chan dstore.QueryWatchEvent, config.DiceConfig.Performance.WatchChanBufSize)
	shardManager := shard.NewShardManager(1, watchChan, nil, globalErrChannel)
	queryWatcherLocal := querymanager.NewQueryManager()
	config.HTTPPort = opt.Port
	// Initialize the HTTPServer
	testServer := server.NewHTTPServer(shardManager)
	// Inform the user that the server is starting
	fmt.Println("Starting the test server on port", config.HTTPPort)
	shardManagerCtx, cancelShardManager := context.WithCancel(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		shardManager.Run(shardManagerCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		queryWatcherLocal.Run(ctx, watchChan)
	}()

	// Start the server in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := testServer.Run(ctx)
		if err != nil {
			cancelShardManager()
			if errors.Is(err, derrors.ErrAborted) {
				return
			}
			if err.Error() != "http: Server closed" {
				log.Fatalf("Http test server encountered an error: %v", err)
			}
			log.Printf("Http test server encountered an error: %v", err)
		}
	}()
}
