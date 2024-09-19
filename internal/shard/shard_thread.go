package shard

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"

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

}

// NewShardThread creates a new ShardThread instance with the given shard id and error channel.
func NewShardThread(id ShardID, errorChan chan *ShardError, watchChan chan dstore.WatchEvent) *ShardThread {
	return &ShardThread{
		id:               id,
		store:            dstore.NewStore(watchChan),
		ReqChan:          make(chan *ops.StoreOp, 1000),
		workerMap:        make(map[string]chan *ops.StoreResponse),
		errorChan:        errorChan,
		lastCronExecTime: utils.GetCurrentTime(),
		cronFrequency:    config.DiceConfig.Server.ShardCronFrequency,
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
	resp := eval.ExecuteCommand(op.Cmd, op.Client, shard.store, op.HTTPOp)

	// resp := shard.executeCommand(op)

	shard.workerMutex.RLock()
	workerChan, ok := shard.workerMap[op.WorkerID]
	shard.workerMutex.RUnlock()

	if ok {
		workerChan <- &ops.StoreResponse{
			RequestID:    op.RequestID,
			EvalResponse: resp,
		}
	} else {
		shard.errorChan <- &ShardError{shardID: shard.id, err: fmt.Errorf(diceerrors.WorkerNotFoundErr, op.WorkerID)}
	}
}

func (shard *ShardThread) executeCommand(op *ops.StoreOp) eval.EvalResponse {

	// Temporary logic till we move all commands to new eval logic.
	// eval.NewDiceCmds map contains refactored eval commands
	// For any command we will first check in the exisiting map
	// if command is NA then we will check in the new map
	var name string
	var newdiceCmd eval.NewDiceCmdMeta
	diceCmd, ok := eval.DiceCmds[op.Cmd.Cmd]
	name = diceCmd.Name
	if !ok {
		newdiceCmd, ok = eval.NewDiceCmds[op.Cmd.Cmd]
		if !ok {

			return eval.EvalResponse{Result: nil, Error: fmt.Errorf("unknown command '%s', with args beginning with: %s", op.Cmd.Cmd, strings.Join(op.Cmd.Args, " "))}
		}
		name = newdiceCmd.Name
	}

	// Till the time we refactor to handle QWATCH differently using HTTP Streaming/SSE
	if op.HTTPOp {
		return eval.EvalResponse{Result: diceCmd.Eval(op.Cmd.Args, shard.store), Error: nil}
	}

	// The following commands could be handled at the shard level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	switch name {
	// new implementation for ping command after rewriting eval
	case "PING", "SET":
		return newdiceCmd.Eval(op.Cmd.Args, shard.store)

	// Old implementation kept as it is, but we will be moving
	// to the new implmentation as PING soon for all commands
	case "SUBSCRIBE", "QWATCH":
		return eval.EvalResponse{Result: eval.EvalQWATCH(op.Cmd.Args, op.Client.Fd, shard.store), Error: nil}
	case "UNSUBSCRIBE", "QUNWATCH":
		return eval.EvalResponse{Result: eval.EvalQUNWATCH(op.Cmd.Args, op.Client.Fd), Error: nil}
	case auth.AuthCmd:
		return eval.EvalResponse{Result: eval.EvalAUTH(op.Cmd.Args, op.Client), Error: nil}
	case "ABORT":
		return eval.EvalResponse{Result: clientio.RespOK, Error: nil}
	default:
		return eval.EvalResponse{Result: diceCmd.Eval(op.Cmd.Args, shard.store), Error: nil}
	}
}

// cleanup handles cleanup logic when the shard stops.
func (shard *ShardThread) cleanup() {
	close(shard.ReqChan)
	if config.DiceConfig.Server.WriteAOFOnCleanup {
		// Avoiding AOF dump for test enabled environments as
		// the tests were taking longer due to background tasks which exceeded the WaitDelay,
		// thus causing the test process to be forcibly terminated.
		log.Info("Skipping AOF dump as test env enabled.")
		return
	}

	eval.EvalBGREWRITEAOF([]string{}, shard.store)
}
