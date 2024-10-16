package worker

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"
)

// preProcessRename prepares the RENAME command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeRename function to delete the old key and set the new key.
func preProcessRename(w *BaseWorker, diceDBCmd *cmd.DiceDBCmd) {
	var rc chan *ops.StoreOp
	var sid shard.ShardID

	key := diceDBCmd.Args[0]
	sid, rc = w.shardManager.GetShardInfo(key)

	preCmd := cmd.DiceDBCmd{
		RequestID:        diceDBCmd.RequestID,
		Cmd:              "GET",
		Args:             []string{key},
		PreProcessingReq: true,
	}

	rc <- &ops.StoreOp{
		SeqID:     0,
		RequestID: preCmd.RequestID,
		Cmd:       &preCmd,
		WorkerID:  w.id,
		ShardID:   sid,
		Client:    nil,
	}
}

// preProcessCopy prepares the COPY command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeCopy function to copy the value to the destination key.
func preProcessCopy(w *BaseWorker, diceDBCmd *cmd.DiceDBCmd) {
	var rc chan *ops.StoreOp
	var sid shard.ShardID

	key := diceDBCmd.Args[0]
	sid, rc = w.shardManager.GetShardInfo(key)

	preCmd := cmd.DiceDBCmd{
		RequestID:        diceDBCmd.RequestID,
		Cmd:              "GET",
		Args:             []string{key},
		PreProcessingReq: true,
	}

	rc <- &ops.StoreOp{
		SeqID:     0,
		RequestID: preCmd.RequestID,
		Cmd:       &preCmd,
		WorkerID:  w.id,
		ShardID:   sid,
		Client:    nil,
	}
}
