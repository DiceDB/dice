// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"log/slog"
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dicedb-go/wire"
)

type WAL interface {
	// Init initializes the WAL.
	Init() error
	// Close closes the WAL.
	Close() error
	// LogCommand logs a command to the WAL.
	LogCommand(c *wire.Command) error
	// Replay replays the WAL.
	ReplayCommand(cb func(c *wire.Command) error) error
}

var (
	rotTicker *time.Ticker
	stopCh    chan struct{}
	mu        sync.Mutex
)

var DefaultWAL WAL

func init() {
	stopCh = make(chan struct{})
}

func rotateWAL() {
	mu.Lock()
	defer mu.Unlock()

	if err := DefaultWAL.Close(); err != nil {
		slog.Warn("error closing the WAL", slog.Any("error", err))
	}

	if err := DefaultWAL.Init(); err != nil {
		slog.Warn("error creating a new WAL", slog.Any("error", err))
	}
}

func periodicRotate() {
	for {
		select {
		case <-rotTicker.C:
			rotateWAL()
		case <-stopCh:
			return
		}
	}
}

func startAsyncJobs() {
	go periodicRotate()
}

// TeardownWAL stops the WAL and closes the WAL instance.
func TeardownWAL() {
	close(stopCh)
	rotTicker.Stop()
}

// SetupWAL initializes the WAL based on the configuration.
// It creates a new WAL instance based on the WAL variant and initializes it.
// If the initialization fails, it panics.
func SetupWAL() {
	switch config.Config.WALVariant {
	case "forge":
		DefaultWAL = newWalForge()
	default:
		return
	}

	rotTicker = time.NewTicker(time.Duration(config.Config.WALRotationTimeSec) * time.Second)
	if err := DefaultWAL.Init(); err != nil {
		slog.Error("could not initialize WAL", slog.Any("error", err))
		panic(err)
	}
}
