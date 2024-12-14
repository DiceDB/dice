// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package wal

import (
	"fmt"
	"log/slog"
	sync "sync"
	"time"

	"github.com/dicedb/dice/internal/cmd"
)

type AbstractWAL interface {
	LogCommand(c *cmd.DiceDBCmd)
	Close() error
	Init(t time.Time) error
	ForEachCommand(f func(c cmd.DiceDBCmd) error) error
}

var (
	ticker *time.Ticker
	stopCh chan struct{}
	mu     sync.Mutex
)

func init() {
	ticker = time.NewTicker(1 * time.Minute)
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

func ReplayWAL(wl AbstractWAL) {
	err := wl.ForEachCommand(func(c cmd.DiceDBCmd) error {
		fmt.Println("replaying", c.Cmd, c.Args)
		return nil
	})

	if err != nil {
		slog.Warn("error replaying WAL", slog.Any("error", err))
	}
}
