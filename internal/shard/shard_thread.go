package shard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
)

type ShardID = uint8

type ShardError struct {
	ShardID ShardID // ShardID is the ID of the shard that encountered the error
	Error   error   // Error is the error that occurred
}

// IOChannels holds the communication channels for an io-thread.
// It contains both the common response channel and the preprocessing response channel.
type IOChannels struct {
	CommonResponseChan        chan *ops.StoreResponse // CommonResponseChan is used to send standard responses for io-thread operations.
	PreProcessingResponseChan chan *ops.StoreResponse // PreProcessingResponseChan is used to send responses related to preprocessing operations.
}

type ShardThread struct {
	id               ShardID               // id is the unique identifier for the shard.
	store            *dstore.Store         // store that the shard is responsible for.
	ReqChan          chan *ops.StoreOp     // ReqChan is this shard's channel for receiving requests.
	ioThreadMap      map[string]IOChannels // ioThreadMap maps each io-thread id to its corresponding IOChannels, containing both the common and preprocessing response channels.
	mu               sync.RWMutex          // mu is the ioThreadMap's mutex for thread safety.
	globalErrorChan  chan error            // globalErrorChan is the channel for sending system-level errors.
	shardErrorChan   chan *ShardError      // ShardErrorChan is the channel for sending shard-level errors.
	lastCronExecTime time.Time             // lastCronExecTime is the last time the shard executed cron tasks.
	cronFrequency    time.Duration         // cronFrequency is the frequency at which the shard executes cron tasks.
}

// NewShardThread creates a new ShardThread instance with the given shard id and error channel.
func NewShardThread(id ShardID, gec chan error, sec chan *ShardError, queryWatchChan chan dstore.QueryWatchEvent,
	cmdWatchChan chan dstore.CmdWatchEvent, evictionStrategy dstore.EvictionStrategy) *ShardThread {
	return &ShardThread{
		id:               id,
		store:            dstore.NewStore(queryWatchChan, cmdWatchChan, evictionStrategy),
		ReqChan:          make(chan *ops.StoreOp, 1000),
		ioThreadMap:      make(map[string]IOChannels),
		globalErrorChan:  gec,
		shardErrorChan:   sec,
		lastCronExecTime: utils.GetCurrentTime(),
		cronFrequency:    config.DiceConfig.Performance.ShardCronFrequency,
	}
}

// Start starts the shard thread, listening for incoming requests.
func (shard *ShardThread) Start(ctx context.Context) {
	ticker := time.NewTicker(shard.cronFrequency)
	defer ticker.Stop()

	for {
		select {
		case op := <-shard.ReqChan:
			shard.processRequest(op)
		case <-ticker.C:
			shard.runCronTasks()
		case <-ctx.Done():
			shard.cleanup()
			return
		}
	}
}

// runCronTasks runs the cron tasks for the shard. This includes deleting expired keys.
func (shard *ShardThread) runCronTasks() {
	dstore.DeleteExpiredKeys(shard.store)
	shard.lastCronExecTime = utils.GetCurrentTime()
}

func (shard *ShardThread) registerIOThread(id string, responseChan, preprocessingChan chan *ops.StoreResponse) {
	shard.mu.Lock()
	shard.ioThreadMap[id] = IOChannels{
		CommonResponseChan:        responseChan,
		PreProcessingResponseChan: preprocessingChan,
	}

	shard.mu.Unlock()
}

func (shard *ShardThread) unregisterIOThread(id string) {
	shard.mu.Lock()
	delete(shard.ioThreadMap, id)
	shard.mu.Unlock()
}

// processRequest processes a Store operation for the shard.
func (shard *ShardThread) processRequest(op *ops.StoreOp) {
	shard.mu.RLock()
	ioChannels, ok := shard.ioThreadMap[op.IOThreadID]
	shard.mu.RUnlock()

	ioThreadChan := ioChannels.CommonResponseChan
	preProcessChan := ioChannels.PreProcessingResponseChan

	sp := &ops.StoreResponse{
		RequestID: op.RequestID,
		SeqID:     op.SeqID,
	}

	e := eval.NewEval(op.Cmd, op.Client, shard.store, op.HTTPOp, op.WebsocketOp, op.PreProcessing)

	if op.PreProcessing {
		resp := e.PreProcessCommand()
		sp.EvalResponse = resp
		preProcessChan <- sp
		return
	}

	resp := e.ExecuteCommand()
	if ok {
		sp.EvalResponse = resp
	} else {
		shard.shardErrorChan <- &ShardError{
			ShardID: shard.id,
			Error:   fmt.Errorf(diceerrors.IOThreadNotFoundErr, op.IOThreadID),
		}
	}

	ioThreadChan <- sp
}

// cleanup handles cleanup logic when the shard stops.
func (shard *ShardThread) cleanup() {
	close(shard.ReqChan)
	if !config.DiceConfig.Persistence.Enabled || !config.DiceConfig.Persistence.WriteAOFOnCleanup {
		return
	}
}
