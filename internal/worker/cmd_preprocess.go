package worker

import (
	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
)

// preProcessRename prepares the RENAME command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeRename function to delete the old key and set the new key.
func preProcessRename(w *BaseWorker, diceDBCmd *cmd.DiceDBCmd) error {
	if len(diceDBCmd.Args) < 2 {
		return diceerrors.ErrWrongArgumentCount("COPY")
	}

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

	return nil
}

// preProcessCopy prepares the COPY command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeCopy function to copy the value to the destination key.
func customProcessCopy(w *BaseWorker, diceDBCmd *cmd.DiceDBCmd) error {
	if len(diceDBCmd.Args) < 2 {
		return diceerrors.ErrWrongArgumentCount("COPY")
	}

	sid1, rc1 := w.shardManager.GetShardInfo(diceDBCmd.Args[0])
	sid2, rc2 := w.shardManager.GetShardInfo(diceDBCmd.Args[1])

	preCmdk1 := cmd.DiceDBCmd{
		Cmd:  "COPY",
		Args: []string{diceDBCmd.Args[0]},
	}

	preCmdk2 := cmd.DiceDBCmd{
		Cmd:  "COPY",
		Args: []string{diceDBCmd.Args[1]},
	}

	// Need to get response from both keys to handle Replace or not
	rc1 <- &ops.StoreOp{
		SeqID:         0,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmdk1,
		WorkerID:      w.id,
		ShardID:       sid1,
		Client:        nil,
		PreProcessing: true,
	}

	rc2 <- &ops.StoreOp{
		SeqID:         1,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmdk2,
		WorkerID:      w.id,
		ShardID:       sid2,
		Client:        nil,
		PreProcessing: true,
	}

	return nil
}
