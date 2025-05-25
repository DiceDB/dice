// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"log/slog"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dicedb-go/wire"
)

type WAL interface {
	// Init initializes the WAL.
	// The WAL implementation should start all the background jobs and initialize the WAL.
	Init() error
	// Stop stops the WAL.
	// The WAL implementation should stop all the background jobs and close the WAL.
	Stop()
	// LogCommand logs a command to the WAL.
	LogCommand(c *wire.Command) error
	// Replay replays the command from the WAL.
	ReplayCommand(cb func(c *wire.Command) error) error
}

var DefaultWAL WAL
var (
	stopCh chan struct{}
)

func init() {
	stopCh = make(chan struct{})
}

// TeardownWAL stops the WAL and closes the WAL instance.
func TeardownWAL() {
	close(stopCh)
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

	if err := DefaultWAL.Init(); err != nil {
		slog.Error("could not initialize WAL", slog.Any("error", err))
		panic(err)
	}
}
