package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dicedb/dice/core/auth"
	"github.com/dicedb/dice/core/ops"
	"github.com/dicedb/dice/internal/constants"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core/diceerrors"
	"github.com/dicedb/dice/server/utils"
)

type ShardID int8

type ShardError struct {
	shardID ShardID // shardID is the ID of the shard that encountered the error
	err     error   // err is the error that occurred
}

type ShardThread struct {
	id               ShardID                            // id is the unique identifier for the shard.
	store            *Store                             // store that the shard is responsible for.
	ReqChan          chan *ops.StoreOp                  // ReqChan is this shard's channel for receiving requests.
	workerMap        map[string]chan *ops.StoreResponse // workerMap maps workerID to its unique response channel
	workerMutex      sync.RWMutex                       // workerMutex is the workerMap's mutex for thread safety.
	errorChan        chan *ShardError                   // errorChan is the channel for sending system-level errors.
	lastCronExecTime time.Time                          // lastCronExecTime is the last time the shard executed cron tasks.
	cronFrequency    time.Duration                      // cronFrequency is the frequency at which the shard executes cron tasks.
}

// NewShardThread creates a new ShardThread instance with the given shard id and error channel.
func NewShardThread(id ShardID, errorChan chan *ShardError) *ShardThread {
	return &ShardThread{
		id:               id,
		store:            NewStore(),
		ReqChan:          make(chan *ops.StoreOp, 1000),
		workerMap:        make(map[string]chan *ops.StoreResponse),
		errorChan:        errorChan,
		lastCronExecTime: utils.GetCurrentTime(),
		cronFrequency:    config.ShardCronFrequency,
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
	DeleteExpiredKeys(shard.store)
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
	response := shard.executeCommand(op)

	shard.workerMutex.RLock()
	workerChan, ok := shard.workerMap[op.WorkerID]
	shard.workerMutex.RUnlock()

	if ok {
		workerChan <- &ops.StoreResponse{
			RequestID: op.RequestID,
			Result:    response,
		}
	} else {
		shard.errorChan <- &ShardError{shardID: shard.id, err: fmt.Errorf(diceerrors.WorkerNotFoundErr, op.WorkerID)}
	}
}

func (shard *ShardThread) executeCommand(op *ops.StoreOp) []byte {
	diceCmd, ok := diceCmds[op.Cmd.Cmd]
	if !ok {
		// Temporary workaround for redis tcl commands
		if _, exists := constants.IgnoreCommands[strings.ToUpper(args[0])]; exists {
			return RespOK
		}

		return diceerrors.NewErrWithFormattedMessage("unknown command '%s', with args beginning with: %s", op.Cmd.Cmd, strings.Join(op.Cmd.Args, " "))
	}

	// The following commands could be handled at the server level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	switch diceCmd.Name {
	case "SUBSCRIBE", "QWATCH":
		return evalQWATCH(op.Cmd.Args, op.Client.Fd, shard.store)
	case "UNSUBSCRIBE", "QUNWATCH":
		return evalQUNWATCH(op.Cmd.Args, op.Client.Fd)
	case auth.AuthCmd:
		return evalAUTH(op.Cmd.Args, op.Client)
	case "ABORT":
		return RespOK
	default:
		return diceCmd.Eval(op.Cmd.Args, shard.store)
	}
}

// cleanup handles cleanup logic when the shard stops.
func (shard *ShardThread) cleanup() {
	close(shard.ReqChan)
	evalBGREWRITEAOF([]string{}, shard.store)
}
