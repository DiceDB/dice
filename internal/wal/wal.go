// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"log/slog"
	"sync"
	"time"

	w "github.com/dicedb/dicedb-go/wal"
	"github.com/dicedb/dicedb-go/wire"
)

type WAL interface {
	Init(t time.Time) error
	LogCommand(c *wire.Command) error
	Close() error
	Replay(c func(*w.Element) error) error
	Iterate(e *w.Element, c func(*w.Element) error) error
}

var (
	ticker *time.Ticker
	stopCh chan struct{}
	mu     sync.Mutex
)

var DefaultWAL WAL

func init() {
	ticker = time.NewTicker(10 * time.Second)
	stopCh = make(chan struct{})
}

func rotateWAL() {
	mu.Lock()
	defer mu.Unlock()

	if err := DefaultWAL.Close(); err != nil {
		slog.Warn("error closing the WAL", slog.Any("error", err))
	}

	if err := DefaultWAL.Init(time.Now()); err != nil {
		slog.Warn("error creating a new WAL", slog.Any("error", err))
	}
}

func periodicRotate() {
	for {
		select {
		case <-ticker.C:
			rotateWAL()
		case <-stopCh:
			return
		}
	}
}

func RunAsyncJobs() {
	go periodicRotate()
}

func ShutdownBG() {
	close(stopCh)
	ticker.Stop()
}
