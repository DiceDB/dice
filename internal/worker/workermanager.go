package worker

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/dicedb/dice/internal/shard"
)

type WorkerManager struct {
	connectedClients sync.Map
	numWorkers       atomic.Int32
	maxClients       int32
	shardManager     *shard.ShardManager
	mu               sync.Mutex
}

var (
	ErrMaxClientsReached = errors.New("maximum number of clients reached")
	ErrWorkerNotFound    = errors.New("worker not found")
)

func NewWorkerManager(maxClients int32, sm *shard.ShardManager) *WorkerManager {
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
	responseChan := worker.(*BaseWorker).responseChan
	preprocessingChan := worker.(*BaseWorker).preprocessingChan

	if responseChan != nil && preprocessingChan != nil {
		wm.shardManager.RegisterWorker(worker.ID(), responseChan, preprocessingChan) // TODO: Change responseChan type to ShardResponse
	} else if responseChan != nil && preprocessingChan == nil {
		wm.shardManager.RegisterWorker(worker.ID(), responseChan, nil)
	}

	wm.numWorkers.Add(1)
	return nil
}

func (wm *WorkerManager) GetWorkerCount() int32 {
	return wm.numWorkers.Load()
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
	wm.numWorkers.Add(-1)

	return nil
}
