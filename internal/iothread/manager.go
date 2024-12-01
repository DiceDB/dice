package iothread

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/dicedb/dice/internal/shard"
)

type Manager struct {
	connectedClients sync.Map
	numIOThreads     atomic.Int32
	maxClients       int32
	shardManager     *shard.ShardManager
	mu               sync.Mutex
}

var (
	ErrMaxClientsReached = errors.New("maximum number of clients reached")
	ErrIOThreadNotFound  = errors.New("io-thread not found")
)

func NewManager(maxClients int32, sm *shard.ShardManager) *Manager {
	return &Manager{
		maxClients:   maxClients,
		shardManager: sm,
	}
}

func (m *Manager) RegisterIOThread(ioThread IOThread) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IOThreadCount() >= m.maxClients {
		return ErrMaxClientsReached
	}

	m.connectedClients.Store(ioThread.ID(), ioThread)
	responseChan := ioThread.(*BaseIOThread).responseChan
	preprocessingChan := ioThread.(*BaseIOThread).preprocessingChan

	if responseChan != nil && preprocessingChan != nil {
		m.shardManager.RegisterIOThread(ioThread.ID(), responseChan, preprocessingChan) // TODO: Change responseChan type to ShardResponse
	} else if responseChan != nil && preprocessingChan == nil {
		m.shardManager.RegisterIOThread(ioThread.ID(), responseChan, nil)
	}

	m.numIOThreads.Add(1)
	return nil
}

func (m *Manager) IOThreadCount() int32 {
	return m.numIOThreads.Load()
}

func (m *Manager) GetIOThread(id string) (IOThread, bool) {
	client, ok := m.connectedClients.Load(id)
	if !ok {
		return nil, false
	}
	return client.(IOThread), true
}

func (m *Manager) UnregisterIOThread(id string) error {
	if client, loaded := m.connectedClients.LoadAndDelete(id); loaded {
		w := client.(IOThread)
		if err := w.Stop(); err != nil {
			return err
		}
	} else {
		return ErrIOThreadNotFound
	}

	m.shardManager.UnregisterIOThread(id)
	m.numIOThreads.Add(-1)

	return nil
}
