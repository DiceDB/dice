// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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

	"github.com/dicedb/dice/internal/server/httpws"

	"github.com/dicedb/dice/config"
	derrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shard"
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

func init() {
	parser := config.NewConfigParser()
	if err := parser.ParseDefaults(config.DiceConfig); err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
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

func (cmd HTTPCommand) IsEmptyCommand() bool {
	return cmd.Command == ""
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
		var result httpws.HTTPResponse
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
	shardManager := shard.NewShardManager(1, nil, globalErrChannel)

	config.DiceConfig.HTTP.Port = opt.Port
	// Initialize the HTTPServer
	testServer := httpws.NewHTTPServer(shardManager, nil)
	// Inform the user that the server is starting
	fmt.Println("Starting the test server on port", config.DiceConfig.HTTP.Port)
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
