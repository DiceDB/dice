package eval

import (
	"sync"

	"github.com/dicedb/dice/internal/global"
)

var (
	clients = make([]global.IOThread, 0)
	mu      = sync.Mutex{}
)

func GetClients() []global.IOThread {
	return clients
}

func AddClient(client global.IOThread) {
	mu.Lock()
	defer mu.Unlock()
	clients = append(clients, client)
}

// if id is monotonic binary search can be used
func RemoveClientByID(id string) {
	mu.Lock()
	defer mu.Unlock()
	for i, client := range clients {
		if client.ID() == id {
			clients = append(clients[:i], clients[i+1:]...)
			return
		}
	}
}
