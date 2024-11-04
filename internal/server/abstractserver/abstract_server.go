package abstractserver

import (
	"context"

	"github.com/dicedb/dice/internal/comm"
)

// move to `comm` pkg?
// contention?
var Clients []*comm.Client = make([]*comm.Client, 0)

type AbstractServer interface {
	Run(ctx context.Context) error
}
