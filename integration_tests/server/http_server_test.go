package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/integration_tests/commands"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

var testHttpOptions = commands.TestServerOptions{
	Port: 8083,
}

func TestHttpServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Cleanup(cancel)

	var wg sync.WaitGroup
	commands.RunHttpServer(ctx, &wg, testHttpOptions)

	time.Sleep(2 * time.Second)

	t.Run("Health Check", func(t *testing.T) {
		t.Parallel()
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", testHttpOptions.Port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Server failed to start on port %d", config.HTTPPort)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		resp.Body.Close()

		if string(body) != "ok" {
			t.Fatalf("Server failed to start on port %d", config.HTTPPort)
		}
	})

	t.Run("ShutdownOnAbort", func(t *testing.T) {
		t.Parallel()
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", testHttpOptions.Port))
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Server failed to start on port %d", config.HTTPPort)
		}

		if err := abort(); err != nil {
			t.Fatalf("Failed to abort server: %v", err)
		}

		time.Sleep(1 * time.Second)

		_, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", testHttpOptions.Port))
		if err != nil {
			t.Fatalf("Http Server failed to shutdown on abort: %v", err)
		}
		resp.Body.Close()
	})
}

func TestHttpServer_HandlesSupportedCommands(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Cleanup(cancel)

	var wg sync.WaitGroup
	commands.RunHttpServer(ctx, &wg, testHttpOptions)

	time.Sleep(2 * time.Second)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", testHttpOptions.Port))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Server failed to start on port %d", config.HTTPPort)
	}

	command := "set"
	key := "What is the name of the first reactive db in the world?"
	value := "dice"
	response, err := makeSetCmdRequest(key, value)
	if err != nil {
		t.Fatalf("failed to handle %s command, err: %v", command, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("failed to handle %s command", command)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "\"OK\"" {
		t.Fatalf("Http failed to handle %s command.", command)
	}

	// to actually test it was successful, lets make a get request
	command2 := "get"
	response2, err := makeGetCmdRequest(key)
	if err != nil {
		t.Fatalf("failed to handle %s command, err: %v", command2, err)
	}
	defer response2.Body.Close()

	getBody, err := io.ReadAll(response2.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(getBody) != fmt.Sprintf("\"%s\"", value) {
		t.Fatalf("server failed to handle %s command, expected (%s), got(%s)", command, value, string(getBody))
	}
}

func makeGetCmdRequest(key string) (*http.Response, error) {
	reqBody := strings.NewReader(fmt.Sprintf(`{"key": "%s"}`, key))
	response, err := http.Post(fmt.Sprintf("http://localhost:%d/get", testHttpOptions.Port), "application/json", reqBody)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func makeSetCmdRequest(key, value string) (*http.Response, error) {
	reqBody := strings.NewReader(fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value))
	response, err := http.Post(fmt.Sprintf("http://localhost:%d/set", testHttpOptions.Port), "application/json", reqBody)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func abort() error {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/abort", testHttpOptions.Port))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("server failed to abort")
	}
	return nil
}
