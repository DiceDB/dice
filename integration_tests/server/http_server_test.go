package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/integration_tests/commands"
	"github.com/dicedb/dice/config"
)

var httpServerOptions = commands.TestServerOptions{
	Port: 8082,
}

func TestHTTPServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Cleanup(cancel)

	var wg sync.WaitGroup
	commands.RunTestServer(ctx, &wg, httpServerOptions)

	time.Sleep(2 * time.Second) // Wait for server to start

	t.Run("HTTPGetRequest", func(t *testing.T) {
		url := fmt.Sprintf("http://localhost:%d/", config.HTTPPort)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to send GET request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		var result interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// You might want to add more specific assertions based on your expected response
	})

	t.Run("HTTPPostRequest", func(t *testing.T) {
		url := fmt.Sprintf("http://localhost:%d/", config.HTTPPort)
		payload := strings.NewReader(`{"cmd": "SET", "args": ["testkey", "testvalue"]}`)
		resp, err := http.Post(url, "application/json", payload)
		if err != nil {
			t.Fatalf("Failed to send POST request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		var result interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Add assertions based on the expected response for a SET command
	})

	t.Run("HealthCheck", func(t *testing.T) {
		url := fmt.Sprintf("http://localhost:%d/health", config.HTTPPort)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to send health check request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code for health check: %d", resp.StatusCode)
		}

		body := make([]byte, 2)
		_, err = resp.Body.Read(body)
		if err != nil || string(body) != "ok" {
			t.Fatalf("Unexpected health check response: %s", string(body))
		}
	})

	wg.Wait()
}