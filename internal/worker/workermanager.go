package worker

import (
	"errors"
	"sync"

	"github.com/dicedb/dice/internal/shard"
)

var (
	ErrInvalidIPAddress = errors.New("invalid IP address")
)

type WorkerManager struct {
	connectedClients sync.Map
	numWorkers       int
	maxClients       int
	shardManager     *shard.ShardManager
	mu               sync.Mutex
}

var (
	ErrMaxClientsReached = errors.New("maximum number of clients reached")
	ErrWorkerNotFound    = errors.New("worker not found")
)

func NewWorkerManager(maxClients int, sm *shard.ShardManager) *WorkerManager {
	return &WorkerManager{
		maxClients:   maxClients,
		shardManager: sm,
	}
}

func (wm *WorkerManager) RegisterWorker(worker Worker) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.GetWorkerCount() >= wm.maxClients {
		return ErrMaxClientsReached
	}

	wm.connectedClients.Store(worker.ID(), worker)
	respChan := worker.(*BaseWorker).respChan
	if respChan != nil {
		wm.shardManager.RegisterWorker(worker.ID(), respChan) // TODO: Change respChan type to ShardResponse
	}

	wm.numWorkers++
	return nil
}

func (wm *WorkerManager) GetWorkerCount() int {
	return wm.numWorkers
}

func (wm *WorkerManager) GetWorker(workerID string) (Worker, bool) {
	worker, ok := wm.connectedClients.Load(workerID)
	if !ok {
		return nil, false
	}
	return worker.(Worker), true
}

func (wm *WorkerManager) UnregisterWorker(workerID string) error {
	if worker, loaded := wm.connectedClients.LoadAndDelete(workerID); loaded {
		w := worker.(Worker)
		if err := w.Stop(); err != nil {
			return err
		}
	} else {
		return ErrWorkerNotFound
	}

	wm.shardManager.UnregisterWorker(workerID)
	wm.numWorkers++

	return nil
}
