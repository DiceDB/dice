// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dicedb/dicedb-go"

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

	go watchLeaderboard()
	go updateScores()

	// Serve static files for the frontend
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("leaderboard running on http://localhost:8000, please open it in your favourite browser.")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func updateScores() {
	ctx := context.Background()
	key := "match:100"
	for {
		dice.ZAdd(ctx, key, dicedb.Z{
			Score:  rand.Float64() * 100,
			Member: fmt.Sprintf("player:%d", rand.Intn(5)),
		})
		time.Sleep(2 * time.Second)
	}
}

func watchLeaderboard() {
	ctx := context.Background()
	watchConn := dice.WatchConn(ctx)
	key := "match:100"
	_, err := watchConn.ZRangeWatch(ctx, key, "0", "4", "REV", "WITHSCORES")
	if err != nil {
		log.Println("failed to create watch connection:", err)
		return
	}

	defer watchConn.Close()

	ch := watchConn.Channel()
	for {
		select {
		case msg := <-ch:
			var entries []LeaderboardEntry
			for _, dicedbZ := range msg.Data.([]dicedb.Z) {
				entry := LeaderboardEntry{
					Score:     int(dicedbZ.Score),
					PlayerID:  dicedbZ.Member.(string),
					Timestamp: time.Now(),
				}
				entries = append(entries, entry)
			}
			broadcast(entries)
		case <-ctx.Done():
			return
		}
	}
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
