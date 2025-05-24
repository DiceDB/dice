// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"time"

	w "github.com/dicedb/dicedb-go/wal"
	"github.com/dicedb/dicedb-go/wire"
)

type WALNull struct {
}

func NewNullWAL() (*WALNull, error) {
	return &WALNull{}, nil
}

func (w *WALNull) Init(t time.Time) error {
	return nil
}

func (w *WALNull) LogCommand(c *wire.Command) error {
	return nil
}

func (w *WALNull) Close() error {
	return nil
}

func (w *WALNull) Replay(callback func(*w.Element) error) error {
	return nil
}

func (w *WALNull) Iterate(entry *w.Element, callback func(*w.Element) error) error {
	return nil
}
