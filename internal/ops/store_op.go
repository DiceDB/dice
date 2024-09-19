package ops

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/eval"
)

type StoreOp struct {
	SeqID     int16         // SeqID is the sequence id of the operation within a single request (optional, may be used for ordering)
	RequestID int32         // RequestID identifies the request that this StoreOp belongs to
	Cmd       *cmd.RedisCmd // Cmd is the atomic Store command (e.g., GET, SET)
	ShardID   int           // ShardID of the shard on which the Store command will be executed
	WorkerID  string        // WorkerID is the ID of the worker that sent this Store operation
	Client    *comm.Client  // Client that sent this Store operation. TODO: This can potentially replace the WorkerID in the future
	HTTPOp    bool          // HTTPOp is true if this Store operation is a HTTP operation
}

// StoreResponse represents the response of a Store operation.
// Store response depends on the response from the evaluation of
// each command and it will be always in the form of
// combination of interface and error
type StoreResponse struct {
	RequestID    int32 // RequestID that this StoreResponse belongs to
	EvalResponse eval.EvalResponse
}
