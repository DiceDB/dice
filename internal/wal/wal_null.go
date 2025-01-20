// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"time"

	"github.com/dicedb/dice/internal/cmd"
)

type WALNull struct {
}

func NewNullWAL() (*WALNull, error) {
	return &WALNull{}, nil
}

func (w *WALNull) Init(t time.Time) error {
	return nil
}

// LogCommand serializes a WALLogEntry and writes it to the current WAL file.
func (w *WALNull) LogCommand(b []byte) error {
	return nil
}

func (w *WALNull) Close() error {
	return nil
}

func (w *WALNull) ForEachCommand(f func(c cmd.DiceDBCmd) error) error {
	return nil
}
