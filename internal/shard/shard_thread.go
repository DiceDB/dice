package shard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"log/slog"

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

type ShardThread struct {
	id               ShardID                            // id is the unique identifier for the shard.
	store            *dstore.Store                      // store that the shard is responsible for.
	ReqChan          chan *ops.StoreOp                  // ReqChan is this shard's channel for receiving requests.
	workerMap        map[string]chan *ops.StoreResponse // workerMap maps workerID to its unique response channel
	workerMutex      sync.RWMutex                       // workerMutex is the workerMap's mutex for thread safety.
	globalErrorChan  chan error                         // globalErrorChan is the channel for sending system-level errors.
	shardErrorChan   chan *ShardError                   // ShardErrorChan is the channel for sending shard-level errors.
	lastCronExecTime time.Time                          // lastCronExecTime is the last time the shard executed cron tasks.
	cronFrequency    time.Duration                      // cronFrequency is the frequency at which the shard executes cron tasks.
	logger           *slog.Logger                       // logger is the logger for the shard.
}

// NewShardThread creates a new ShardThread instance with the given shard id and error channel.
func NewShardThread(id ShardID, gec chan error, sec chan *ShardError, watchChan chan dstore.QueryWatchEvent, logger *slog.Logger) *ShardThread {
	return &ShardThread{
		id:               id,
		store:            dstore.NewStore(watchChan),
		ReqChan:          make(chan *ops.StoreOp, 1000),
		workerMap:        make(map[string]chan *ops.StoreResponse),
		globalErrorChan:  gec,
		shardErrorChan:   sec,
		lastCronExecTime: utils.GetCurrentTime(),
		cronFrequency:    config.DiceConfig.Server.ShardCronFrequency,
		logger:           logger,
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

func (shard *ShardThread) registerWorker(workerID string, workerChan chan *ops.StoreResponse) {
	shard.workerMutex.Lock()
	shard.workerMap[workerID] = workerChan
	shard.workerMutex.Unlock()
}

func (shard *ShardThread) unregisterWorker(workerID string) {
	shard.workerMutex.Lock()
	delete(shard.workerMap, workerID)
	shard.workerMutex.Unlock()
}

// processRequest processes a Store operation for the shard.
func (shard *ShardThread) processRequest(op *ops.StoreOp) {
	resp := eval.ExecuteCommand(op.Cmd, op.Client, shard.store, op.HTTPOp, op.WebsocketOp)

	shard.workerMutex.RLock()
	workerChan, ok := shard.workerMap[op.WorkerID]
	shard.workerMutex.RUnlock()

	sp := &ops.StoreResponse{
		RequestID: op.RequestID,
	}

	if ok {
		sp.EvalResponse = resp
	} else {
		shard.shardErrorChan <- &ShardError{
			ShardID: shard.id,
			Error:   fmt.Errorf(diceerrors.WorkerNotFoundErr, op.WorkerID),
		}
	}

	workerChan <- sp
}

// cleanup handles cleanup logic when the shard stops.
func (shard *ShardThread) cleanup() {
	close(shard.ReqChan)
	if !config.DiceConfig.Server.WriteAOFOnCleanup {
		slog.Info("Skipping AOF dump.")
		return
	}

	eval.EvalBGREWRITEAOF([]string{}, shard.store)
}
