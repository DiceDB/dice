// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"log/slog"
	sync "sync"
	"time"
)

type AbstractWAL interface {
	Log([]byte) error
	Close() error
	Init(t time.Time) error
	Replay(c func(any) error) error
	Iterate(e any, c func(any) error) error
}

type WALCmdEntry struct {
	Command  string   // The command being executed
	Args     []string // Additional command arguments
	ClientID string
}

var (
	ticker   *time.Ticker
	stopCh   chan struct{}
	mu       sync.Mutex
	instance AbstractWAL
)

// GetWAL returns the global WAL instance
func GetWAL() AbstractWAL {
	mu.Lock()
	defer mu.Unlock()
	return instance
}

// SetGlobalWAL sets the global WAL instance
func SetWAL(wal AbstractWAL) {
	mu.Lock()
	defer mu.Unlock()
	instance = wal
}

func init() {
	ticker = time.NewTicker(10 * time.Second)
	stopCh = make(chan struct{})
}

func rotateWAL(wl AbstractWAL) {
	mu.Lock()
	defer mu.Unlock()

	if err := wl.Close(); err != nil {
		slog.Warn("error closing the WAL", slog.Any("error", err))
	}

	if err := wl.Init(time.Now()); err != nil {
		slog.Warn("error creating a new WAL", slog.Any("error", err))
	}
}

func periodicRotate(wl AbstractWAL) {
	for {
		select {
		case <-ticker.C:
			rotateWAL(wl)
		case <-stopCh:
			return
		}
	}
}

func InitBG(wl AbstractWAL) {
	go periodicRotate(wl)
}

func ShutdownBG() {
	close(stopCh)
	ticker.Stop()
}
