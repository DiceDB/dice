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
func (w *WALNull) LogCommand(c *cmd.DiceDBCmd) {
}

func (w *WALNull) Close() error {
	return nil
}

func (w *WALNull) ForEachCommand(f func(c cmd.DiceDBCmd) error) error {
	return nil
}
