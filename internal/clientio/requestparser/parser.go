package requestparser

import (
	"github.com/dicedb/dice/internal/cmd"
)

type Parser interface {
	Parse(data []byte) ([]*cmd.DiceDBCmd, error)
}
