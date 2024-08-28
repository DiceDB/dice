package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core/diceerrors"
	"github.com/dicedb/dice/server/utils"
)

type ShardID int8

type StoreOp struct {
	SeqID     int16     // sequence id of the operation within a single request (optional, may be used for ordering)
	RequestID int32     // The request ID that this StoreOp belongs to
	Cmd       *RedisCmd // The atomic Store command (e.g., GET, SET)
	ShardID   int       // The ID of the shard on which the Store command will be executed
	WorkerID  string    // The ID of the worker that sent this Store operation
	Client    *Client   // The file descriptor of the client that sent this Store operation. TODO: This can potentially replace the WorkerID in the future
}

// StoreResponse represents the response of a Store operation.
type StoreResponse struct {
	requestID int32  // The request ID that this StoreResponse belongs to
	Result    []byte // The Result of the Store operation, for now the type is set to []byte, but this can change in the future.
}

type ShardError struct {
	shardID ShardID // The ID of the shard that encountered the error
	err     error   // The error that occurred
}

type ShardThread struct {
	id               ShardID                        // unique identifier for the shard.
	store            *Store                         // the store that the shard is responsible for.
	ReqChan          chan *StoreOp                  // channel for receiving requests.
	workerMap        map[string]chan *StoreResponse // map of worker id to its unique response channel (sync.Map for thread safety).
	mu               sync.RWMutex                   // mutex for thread safety.
	errorChan        chan *ShardError               // channel for sending system-level errors.
	lastCronExecTime time.Time                      // the last time the shard executed cron tasks.
	cronFrequency    time.Duration                  // the frequency at which the shard executes cron tasks.
}

// NewShardThread creates a new ShardThread instance with the given shard id and error channel.
func NewShardThread(id ShardID, errorChan chan *ShardError) *ShardThread {
	return &ShardThread{
		id:               id,
		store:            NewStore(),
		ReqChan:          make(chan *StoreOp, 1000),
		workerMap:        make(map[string]chan *StoreResponse),
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

func (shard *ShardThread) registerWorker(workerID string, workerChan chan *StoreResponse) {
	shard.mu.Lock()
	shard.workerMap[workerID] = workerChan
	shard.mu.Unlock()
}

// processRequest processes a Store operation for the shard.
func (shard *ShardThread) processRequest(op *StoreOp) {
	response := shard.executeCommand(op)

	// Send the Result to the worker if worker exists.
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	if workerChan, ok := shard.workerMap[op.WorkerID]; ok {
		workerChan <- &StoreResponse{
			requestID: op.RequestID,
			Result:    response,
		}
	} else {
		shard.errorChan <- &ShardError{shardID: shard.id, err: fmt.Errorf(diceerrors.WorkerNotFoundErr, op.WorkerID)}
	}
}

func (shard *ShardThread) executeCommand(op *StoreOp) []byte {
	diceCmd, ok := diceCmds[op.Cmd.Cmd]
	if !ok {
		return diceerrors.NewErrWithFormattedMessage("unknown command '%s', with args beginning with: %s", op.Cmd.Cmd, strings.Join(op.Cmd.Args, " "))
	}

	// The following commands could be handled at the server level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	if diceCmd.Name == "SUBSCRIBE" || diceCmd.Name == "QWATCH" {
		return evalQWATCH(op.Cmd.Args, op.Client.Fd, shard.store)
	}
	if diceCmd.Name == "UNSUBSCRIBE" || diceCmd.Name == "QUNWATCH" {
		return evalQUNWATCH(op.Cmd.Args, op.Client.Fd)
	}
	if diceCmd.Name == AuthCmd {
		return evalAUTH(op.Cmd.Args, op.Client)
	}
	if diceCmd.Name == "ABORT" {
		return RespOK
	}

	return diceCmd.Eval(op.Cmd.Args, shard.store)
}

// cleanup handles cleanup logic when the shard stops.
func (shard *ShardThread) cleanup() {
	close(shard.ReqChan)
	evalBGREWRITEAOF([]string{}, shard.store)
}
