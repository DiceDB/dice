package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dicedb/go-dice"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	dice    *redis.Client
	upgrade = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all connections for simplicity
		},
	}
)

type LeaderboardEntry struct {
	PlayerID  string    `json:"player_id"`
	Score     int       `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	time.Sleep(2 * time.Second)
	// Initialize Redis client
	dice = redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", os.Getenv("DICEDB_HOST"), os.Getenv("DICEDB_PORT")),
		DialTimeout: 10 * time.Second,
		MaxRetries:  10,
	})

	// Start the leaderboard update goroutine
	go updateLeaderboard()

	// Start the QWATCH listener goroutine
	go watchLeaderboard()

	// Serve static files for the frontend.
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// Set up WebSocket endpoint
	http.HandleFunc("/ws", handleWebSocket)

	// Start the HTTP server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// updateLeaderboard generates random leaderboard entries and updates them in Redis.
func updateLeaderboard() {
	ctx := context.Background()
	for {
		entry := LeaderboardEntry{
			PlayerID:  fmt.Sprintf("player_%d", rand.Intn(100)),
			Score:     rand.Intn(10000),
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(entry)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			continue
		}

		err = dice.JSONSet(ctx, entry.PlayerID, "$", jsonData).Err()
		if err != nil {
			log.Printf("Error setting data in Redis: %v", err)
		}

		// Expire keys after 10 seconds to prevent leaderboard from becoming static after a while.
		err = dice.ExpireAt(ctx, entry.PlayerID, time.Now().Add(10*time.Second)).Err()
		if err != nil {
			log.Printf("Error setting expiration in Redis: %v", err)
		}

		time.Sleep(2 * time.Second)
	}
}

// watchLeaderboard watches the leaderboard for updates and sends them to all connected WebSocket clients.
func watchLeaderboard() {
	ctx := context.Background()
	qwatch := dice.QWatch(ctx)
	err := qwatch.WatchQuery(ctx, "SELECT $key, $value FROM `player_*` WHERE '$value.score' > 1000 ORDER BY $value.score DESC LIMIT 5")
	if err != nil {
		log.Fatalf("Error watching query: %v", err)
	}
	defer qwatch.Close()

	ch := qwatch.Channel()
	for {
		select {
		case msg := <-ch:
			updates, err := formatToJSON(msg.Updates)
			if err != nil {
				log.Printf("Error formatting updates: %v", err)
				continue
			}
			// Broadcast the formatted result to all connected WebSocket clients
			broadcastToClients(updates)
		case <-ctx.Done():
			return
		}
	}
}

////////////////////////
// Helper functions
////////////////////////

func formatToJSON(updates []redis.KV) (string, error) {
	var entries []LeaderboardEntry
	for _, update := range updates {
		var entry LeaderboardEntry
		err := json.Unmarshal([]byte(update.Value.(string)), &entry)
		if err != nil {
			return "", fmt.Errorf("error unmarshaling entry: %v", err)
		}
		entries = append(entries, entry)
	}
	jsonData, err := json.Marshal(entries)
	if err != nil {
		return "", fmt.Errorf("error marshaling entries: %v", err)
	}
	return string(jsonData), nil
}

var (
	clients    = make(map[*websocket.Conn]bool)
	clientsMux = &sync.Mutex{}
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}(conn)

	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()

	// Keep the connection open
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			break
		}
	}

	clientsMux.Lock()
	delete(clients, conn)
	clientsMux.Unlock()
}

func broadcastToClients(message string) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			err := client.Close()
			if err != nil {
				log.Printf("Error closing WebSocket connection: %v", err)
			}
			delete(clients, client)
		}
	}
}
