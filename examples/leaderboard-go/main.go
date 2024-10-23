package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dicedb/dicedb-go"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	dice    *dicedb.Client
	upgrade = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
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

	dhost := "localhost"
	if val := os.Getenv("DICEDB_HOST"); val != "" {
		dhost = val
	}

	dport := "7379"
	if val := os.Getenv("DICEDB_PORT"); val != "" {
		dport = val
	}

	dice = dicedb.NewClient(&dicedb.Options{
		Addr:        fmt.Sprintf("%s:%s", dhost, dport),
		DialTimeout: 10 * time.Second,
		MaxRetries:  10,
	})

	go updateScores()
	go watchLeaderboard()

	// Serve static files for the frontend
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("leaderboard running on http://localhost:8000, please open it in your favourite browser.")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func updateScores() {
	ctx := context.Background()
	for {
		entry := LeaderboardEntry{
			PlayerID:  fmt.Sprintf("player:%d", rand.Intn(10)),
			Score:     rand.Intn(100),
			Timestamp: time.Now(),
		}
		lentry, _ := json.Marshal(entry)
		dice.JSONSet(ctx, entry.PlayerID, "$", lentry).Err()
	}
}

func watchLeaderboard() {
	ctx := context.Background()
	qwatch := dice.QWatch(ctx)
	qwatch.WatchQuery(ctx, `SELECT $key, $value
									WHERE $key LIKE 'player:*' AND '$value.score' > 10
									ORDER BY $value.score DESC
									LIMIT 5;`)
	defer qwatch.Close()

	ch := qwatch.Channel()
	for {
		select {
		case msg := <-ch:
			entries := toEntries(msg.Updates)
			broadcast(entries)
		case <-ctx.Done():
			return
		}
	}
}

func toEntries(updates []dicedb.KV) []LeaderboardEntry {
	var entries []LeaderboardEntry
	for _, update := range updates {
		var entry LeaderboardEntry
		json.Unmarshal([]byte(update.Value.(string)), &entry)
		entries = append(entries, entry)
	}
	return entries
}

func broadcast(entries []LeaderboardEntry) {
	cMux.Lock()
	defer cMux.Unlock()

	message, _ := json.Marshal(entries)
	for client := range clients {
		client.WriteMessage(websocket.TextMessage, message)
	}
}

var (
	clients = make(map[*websocket.Conn]bool)
	cMux    = &sync.Mutex{}
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading to WebSocket: %v", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("error closing WebSocket connection: %v", err)
		}
	}(conn)

	cMux.Lock()
	clients[conn] = true
	cMux.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}

	cMux.Lock()
	delete(clients, conn)
	cMux.Unlock()
}
