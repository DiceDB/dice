package abstractserver

import (
	"context"
	"sync"

	"github.com/dicedb/dice/internal/comm"
)

// move to `comm` pkg?
var (
	clients = make([]*comm.Client, 0)
	mu      = sync.Mutex{}
)

func GetClients() []*comm.Client {
	return clients
}

func AddClient(client *comm.Client) {
	mu.Lock()
	defer mu.Unlock()
	clients = append(clients, client)
}

// if Client.Fd or Client.ID is monotonic, binary search should be used
func RemoveClientByFd(fd int) {
	mu.Lock()
	defer mu.Unlock()
	for i, client := range clients {
		if client.Fd == fd {
			clients = append(clients[:i], clients[i+1:]...)
			return
		}
	}
}

type AbstractServer interface {
	Run(ctx context.Context) error
}
