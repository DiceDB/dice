// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/dicedb/dice/config"
)

type IOThreadManager struct {
	connectedClients sync.Map
	numIOThreads     atomic.Uint32
	mu               sync.Mutex
}

var (
	ErrMaxClientsReached = errors.New("maximum number of clients reached")
	ErrIOThreadNotFound  = errors.New("io-thread not found")
)

func NewIOThreadManager() *IOThreadManager {
	return &IOThreadManager{}
}

func (m *IOThreadManager) RegisterIOThread(ioThread *IOThread) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IOThreadCount() >= uint32(config.Config.MaxClients) {
		return ErrMaxClientsReached
	}

	m.connectedClients.Store(ioThread.ClientID, ioThread)
	m.numIOThreads.Add(1)
	return nil
}

func (m *IOThreadManager) IOThreadCount() uint32 {
	return m.numIOThreads.Load()
}

func (m *IOThreadManager) UnregisterIOThread(id string) error {
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
