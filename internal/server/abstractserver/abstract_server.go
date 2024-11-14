package abstractserver

import (
	"context"

	"github.com/dicedb/dice/internal/comm"
)

// move to `comm` pkg?
// contention?
var Clients []*comm.Client = make([]*comm.Client, 0)

// if Client.Fd or Client.ID is monotonic, binary search should be used
func RemoveClientByFd(fd int) {
	for i, client := range Clients {
		if client.Fd == fd {
			Clients = append(Clients[:i], Clients[i+1:]...)
			return
		}
	}
}

type AbstractServer interface {
	Run(ctx context.Context) error
}
