package iothread

import (
	"errors"
	"sync"
	"sync/atomic"
)

type Manager struct {
	connectedClients sync.Map
	numIOThreads     atomic.Int32
	maxClients       int32
	mu               sync.Mutex
}

var (
	ErrMaxClientsReached = errors.New("maximum number of clients reached")
	ErrIOThreadNotFound  = errors.New("io-thread not found")
)

func NewManager(maxClients int32) *Manager {
	return &Manager{
		maxClients: maxClients,
	}
}

func (m *Manager) RegisterIOThread(ioThread IOThread) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IOThreadCount() >= m.maxClients {
		return ErrMaxClientsReached
	}

	m.connectedClients.Store(ioThread.ID(), ioThread)

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

	m.numIOThreads.Add(-1)
	return nil
}
