// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package commandhandler

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/dicedb/dice/internal/shard"
)

type Registry struct {
	activeCmdHandlers sync.Map
	numCmdHandlers    atomic.Uint32
	maxCmdHandlers    uint32
	ShardManager      *shard.ShardManager
	mu                sync.Mutex
}

var (
	ErrMaxCmdHandlersReached     = errors.New("maximum number of command handlers reached")
	ErrCmdHandlerNotFound        = errors.New("command handler not found")
	ErrCmdHandlerNotBase         = errors.New("command handler is not a BaseCommandHandler")
	ErrCmdHandlerResponseChanNil = errors.New("command handler response channel is nil")
)

func NewRegistry(maxClients uint32, sm *shard.ShardManager) *Registry {
	return &Registry{
		maxCmdHandlers: maxClients,
		ShardManager:   sm,
	}
}

func (m *Registry) RegisterCommandHandler(cmdHandler *BaseCommandHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.CommandHandlerCount() >= m.maxCmdHandlers {
		return ErrMaxCmdHandlersReached
	}

	if cmdHandler.responseChan == nil {
		return ErrCmdHandlerResponseChanNil
	}

	m.ShardManager.RegisterCommandHandler(cmdHandler.ID(), cmdHandler.responseChan, cmdHandler.preprocessingChan)

	m.activeCmdHandlers.Store(cmdHandler.ID(), cmdHandler)

	m.numCmdHandlers.Add(1)
	return nil
}

func (m *Registry) CommandHandlerCount() uint32 {
	return m.numCmdHandlers.Load()
}

func (m *Registry) UnregisterCommandHandler(id string) error {
	m.ShardManager.UnregisterCommandHandler(id)
	if cmdHandler, loaded := m.activeCmdHandlers.LoadAndDelete(id); loaded {
		ch, ok := cmdHandler.(*BaseCommandHandler)
		if !ok {
			return ErrCmdHandlerNotBase
		}
		if err := ch.Stop(); err != nil {
			return err
		}
	} else {
		return ErrCmdHandlerNotFound
	}
	m.numCmdHandlers.Add(^uint32(0))
	return nil
}
