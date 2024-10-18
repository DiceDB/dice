package worker

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/ops"
)

// preProcessRename prepares the RENAME command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeRename function to delete the old key and set the new key.
func preProcessRename(w *BaseWorker, diceDBCmd *cmd.DiceDBCmd) {
	key := diceDBCmd.Args[0]
	sid, rc := w.shardManager.GetShardInfo(key)

	preCmd := cmd.DiceDBCmd{
		Cmd:  CmdGet,
		Args: []string{key},
	}

	rc <- &ops.StoreOp{
		SeqID:         0,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmd,
		WorkerID:      w.id,
		ShardID:       sid,
		Client:        nil,
		PreProcessing: true,
	}
}

// preProcessCopy prepares the COPY command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeCopy function to copy the value to the destination key.
func preProcessCopy(w *BaseWorker, diceDBCmd *cmd.DiceDBCmd) {
	key := diceDBCmd.Args[0]
	sid, rc := w.shardManager.GetShardInfo(key)

	preCmd := cmd.DiceDBCmd{
		Cmd:  CmdGet,
		Args: []string{key},
	}

	rc <- &ops.StoreOp{
		SeqID:         0,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmd,
		WorkerID:      w.id,
		ShardID:       sid,
		Client:        nil,
		PreProcessing: true,
	}
}
