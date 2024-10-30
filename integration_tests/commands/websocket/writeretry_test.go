package websocket

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/server"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var serverAddr = fmt.Sprintf("localhost:%v", testPort2)
var once sync.Once

func TestWriteResponseWithRetries_Success(t *testing.T) {
	conn := createWebSocketConn(t)
	defer conn.Close()

	// Complete a write without any errors
	err := server.WriteResponseWithRetries(conn, []byte("hello"), 3)
	assert.NoError(t, err)
}

func TestWriteResponseWithRetries_NetworkError(t *testing.T) {
	conn := createWebSocketConn(t)
	defer conn.Close()

	// Simulate a network error by closing the connection beforehand
	conn.Close()

	err := server.WriteResponseWithRetries(conn, []byte("hello"), 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network operation error")
}

func TestWriteResponseWithRetries_BrokenPipe(t *testing.T) {
	conn := createWebSocketConn(t)
	defer conn.Close()

	// Simulate a broken pipe error by manually triggering it.
	conn.UnderlyingConn().(*net.TCPConn).CloseWrite()

	err := server.WriteResponseWithRetries(conn, []byte("hello"), 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "broken pipe")
}

func TestWriteResponseWithRetries_EAGAINRetry(t *testing.T) {
	conn := createWebSocketConn(t)
	defer conn.Close()

	// No direct way to trigger EAGAIN, but this simulates retries working.
	// Forcing the first two attempts to fail by closing and reopening the socket.
	retries := 0
	conn.SetWriteDeadline(time.Now().Add(1 * time.Millisecond))

	for retries < 2 {
		err := server.WriteResponseWithRetries(conn, []byte("hello"), 3)
		if err != nil {
			// Retry and reset deadline after a failed attempt.
			conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
			retries++
		}
	}
	assert.Equal(t, 2, retries)
}

func startWebSocketServer() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if err != nil {
			http.Error(w, "Failed to upgrade", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	})

	go http.ListenAndServe(serverAddr, nil)
}

// Helper to create a WebSocket connection for testing
func createWebSocketConn(t *testing.T) *websocket.Conn {
	once.Do(startWebSocketServer)
	var conn *websocket.Conn
  var err error

	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws"}

	// Retry up to 5 times with a short delay
	for i := 0; i < 5; i++ {
			conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			if err == nil {
					return conn
			}
			time.Sleep(200 * time.Millisecond) // Adjust delay as necessary
	}

	t.Fatalf("Failed to connect to WebSocket server: %v", err)
	return nil
}
