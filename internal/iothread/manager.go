// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iothread

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/dicedb/dice/config"
)

type Manager struct {
	connectedClients sync.Map
	numIOThreads     atomic.Uint32
	mu               sync.Mutex
}

var (
	ErrMaxClientsReached = errors.New("maximum number of clients reached")
	ErrIOThreadNotFound  = errors.New("io-thread not found")
)

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) RegisterIOThread(ioThread *IOThread) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IOThreadCount() >= uint32(config.Config.MaxClients) {
		return ErrMaxClientsReached
	}

	m.connectedClients.Store(ioThread.ID(), ioThread)

	m.numIOThreads.Add(1)
	return nil
}

func (m *Manager) IOThreadCount() uint32 {
	return m.numIOThreads.Load()
}

func (m *Manager) UnregisterIOThread(id string) error {
	if client, loaded := m.connectedClients.LoadAndDelete(id); loaded {
		w := client.(*IOThread)
		if err := w.Stop(); err != nil {
			return err
		}
	} else {
		return ErrIOThreadNotFound
	}

	m.numIOThreads.Add(^uint32(0))
	return nil
}
