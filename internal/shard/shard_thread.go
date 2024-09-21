package shard

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
)

type ShardID int8

type ShardError struct {
	shardID ShardID // shardID is the ID of the shard that encountered the error
	err     error   // err is the error that occurred
}

type ShardThread struct {
	id               ShardID                            // id is the unique identifier for the shard.
	store            *dstore.Store                      // store that the shard is responsible for.
	ReqChan          chan *ops.StoreOp                  // ReqChan is this shard's channel for receiving requests.
	workerMap        map[string]chan *ops.StoreResponse // workerMap maps workerID to its unique response channel
	workerMutex      sync.RWMutex                       // workerMutex is the workerMap's mutex for thread safety.
	errorChan        chan *ShardError                   // errorChan is the channel for sending system-level errors.
	lastCronExecTime time.Time                          // lastCronExecTime is the last time the shard executed cron tasks.
	cronFrequency    time.Duration                      // cronFrequency is the frequency at which the shard executes cron tasks.
	logger           *slog.Logger                       // logger is the logger for the shard.
}

// NewShardThread creates a new ShardThread instance with the given shard id and error channel.
func NewShardThread(id ShardID, errorChan chan *ShardError, watchChan chan dstore.WatchEvent, logger *slog.Logger) *ShardThread {
	return &ShardThread{
		id:               id,
		store:            dstore.NewStore(watchChan),
		ReqChan:          make(chan *ops.StoreOp, 1000),
		workerMap:        make(map[string]chan *ops.StoreResponse),
		errorChan:        errorChan,
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
	resp := shard.executeCommand(op)

	shard.workerMutex.RLock()
	workerChan, ok := shard.workerMap[op.WorkerID]
	shard.workerMutex.RUnlock()

	if ok {
		workerChan <- &ops.StoreResponse{
			RequestID: op.RequestID,
			Result:    resp,
		}
	} else {
		shard.errorChan <- &ShardError{shardID: shard.id, err: fmt.Errorf(diceerrors.WorkerNotFoundErr, op.WorkerID)}
	}
}

func (shard *ShardThread) executeCommand(op *ops.StoreOp) []byte {
	diceCmd, ok := eval.DiceCmds[op.Cmd.Cmd]
	if !ok {
		return diceerrors.NewErrWithFormattedMessage("unknown command '%s', with args beginning with: %s", op.Cmd.Cmd, strings.Join(op.Cmd.Args, " "))
	}

	// Till the time we refactor to handle QWATCH differently using HTTP Streaming/SSE
	if op.HTTPOp {
		return diceCmd.Eval(op.Cmd.Args, shard.store)
	}

	// The following commands could be handled at the shard level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	switch diceCmd.Name {
	case "SUBSCRIBE", "QWATCH":
		return eval.EvalQWATCH(op.Cmd.Args, op.Client.Fd, shard.store)
	case "UNSUBSCRIBE", "QUNWATCH":
		return eval.EvalQUNWATCH(op.Cmd.Args, op.Client.Fd)
	case auth.AuthCmd:
		return eval.EvalAUTH(op.Cmd.Args, op.Client)
	case "ABORT":
		return clientio.RespOK
	default:
		return diceCmd.Eval(op.Cmd.Args, shard.store)
	}
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
