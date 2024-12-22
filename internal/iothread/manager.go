// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
