package commandhandler

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/dicedb/dice/internal/shard"
)

type Registry struct {
	activeCmdHandlers sync.Map
	numCmdHandlers    atomic.Int32
	maxClients        int32
	ShardManager      *shard.ShardManager
	mu                sync.Mutex
}

var (
	ErrMaxCmdHandlersReached = errors.New("maximum number of command handlers reached")
	ErrCmdHandlerNotFound    = errors.New("command handler not found")
)

func NewRegistry(maxClients int32, sm *shard.ShardManager) *Registry {
	return &Registry{
		maxClients:   maxClients,
		ShardManager: sm,
	}
}

func (m *Registry) RegisterCommandHandler(cmdHandler *BaseCommandHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.CommandHandlerCount() >= m.maxClients {
		return ErrMaxCmdHandlersReached
	}

	responseChan := cmdHandler.responseChan
	preprocessingChan := cmdHandler.preprocessingChan

	if responseChan != nil && preprocessingChan != nil {
		m.ShardManager.RegisterCommandHandler(cmdHandler.ID(), responseChan, preprocessingChan) // TODO: Change responseChan type to ShardResponse
	} else if responseChan != nil && preprocessingChan == nil {
		m.ShardManager.RegisterCommandHandler(cmdHandler.ID(), responseChan, nil)
	}

	m.activeCmdHandlers.Store(cmdHandler.ID(), cmdHandler)

	m.numCmdHandlers.Add(1)
	return nil
}

func (m *Registry) CommandHandlerCount() int32 {
	return m.numCmdHandlers.Load()
}

func (m *Registry) UnregisterCommandHandler(id string) error {
	m.ShardManager.UnregisterCommandHandler(id)
	if cmdHandler, loaded := m.activeCmdHandlers.LoadAndDelete(id); loaded {
		ch := cmdHandler.(*BaseCommandHandler)
		if err := ch.Stop(); err != nil {
			return err
		}
	} else {
		return ErrCmdHandlerNotFound
	}
	m.numCmdHandlers.Add(-1)
	return nil
}
