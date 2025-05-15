// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"leaderboard-go/svc"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dicedb/dicedb-go/wire"
)

var players = []string{"Alice", "Bob", "Charlie", "Dora", "Evan", "Fay", "Gina"}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mock" {
		for {
			player := players[rand.Intn(len(players))]
			score := rand.Intn(100)
			svc.UpdateScore(player, score)
			fmt.Println("updated game:scores for", player, "with score", score)
			time.Sleep(time.Millisecond * 500)
		}
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Subscribe to leaderboard updates
	svc.Subscribe()

	// Start listening for messages
	go svc.ListenForMessages(func(result *wire.Result) {
		displayLeaderboard(result.GetZRANGERes().Elements)
	})

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nShutting down...")
}

func displayLeaderboard(leaderboard []*wire.ZElement) {
	// Clear the screen
	fmt.Print("\033[H\033[2J")

	fmt.Println("Rank  Score  Player")
	fmt.Println("------------------")

	for _, element := range leaderboard {
		fmt.Printf("%2d.   %4d   %s\n", element.Rank, element.Score, element.Member)
	}

	fmt.Println("------------------")
	fmt.Println("Press Ctrl+C to exit")
}
