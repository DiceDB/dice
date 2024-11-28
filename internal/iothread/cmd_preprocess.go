package iothread

import (
	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
)

// preProcessRename prepares the RENAME command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeRename function to delete the old key and set the new key.
func preProcessRename(thread *BaseIOThread, diceDBCmd *cmd.DiceDBCmd) error {
	if len(diceDBCmd.Args) < 2 {
		return diceerrors.ErrWrongArgumentCount("RENAME")
	}

	key := diceDBCmd.Args[0]
	sid, rc := thread.shardManager.GetShardInfo(key)

	preCmd := cmd.DiceDBCmd{
		Cmd:  "RENAME",
		Args: []string{key},
	}

	rc <- &ops.StoreOp{
		SeqID:         0,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmd,
		IOThreadID:    thread.id,
		ShardID:       sid,
		Client:        nil,
		PreProcessing: true,
	}

	return nil
}

// preProcessCopy prepares the COPY command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeCopy function to copy the value to the destination key.
func customProcessCopy(thread *BaseIOThread, diceDBCmd *cmd.DiceDBCmd) error {
	if len(diceDBCmd.Args) < 2 {
		return diceerrors.ErrWrongArgumentCount("COPY")
	}

	sid, rc := thread.shardManager.GetShardInfo(diceDBCmd.Args[0])

	preCmd := cmd.DiceDBCmd{
		Cmd:  "COPY",
		Args: []string{diceDBCmd.Args[0]},
	}

	// Need to get response from both keys to handle Replace or not
	rc <- &ops.StoreOp{
		SeqID:         0,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmd,
		IOThreadID:    thread.id,
		ShardID:       sid,
		Client:        nil,
		PreProcessing: true,
	}

	return nil
}
