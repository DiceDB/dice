package ops

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/eval"
)

type StoreOp struct {
	SeqID       uint8         // SeqID is the sequence id of the operation within a single request (optional, may be used for ordering)
	RequestID   uint32        // RequestID identifies the request that this StoreOp belongs to
	Cmd         *cmd.RedisCmd // Cmd is the atomic Store command (e.g., GET, SET)
	ShardID     uint8         // ShardID of the shard on which the Store command will be executed
	WorkerID    string        // WorkerID is the ID of the worker that sent this Store operation
	Client      *comm.Client  // Client that sent this Store operation. TODO: This can potentially replace the WorkerID in the future
	HTTPOp      bool          // HTTPOp is true if this Store operation is an HTTP operation
	WebsocketOp bool          // WebsocketOp is true if this Store operation is a Websocket operation
}

// StoreResponse represents the response of a Store operation.
type StoreResponse struct {
	RequestID    uint32            // RequestID that this StoreResponse belongs to
	EvalResponse eval.EvalResponse // Result of the Store operation, for now the type is set to []byte, but this can change in the future.
}
