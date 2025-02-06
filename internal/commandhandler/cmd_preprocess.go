// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package commandhandler

import (
	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
)

// preProcessRename prepares the RENAME command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeRename function to delete the old key and set the new key.
func preProcessRename(h *BaseCommandHandler, diceDBCmd *cmd.DiceDBCmd) error {
	if len(diceDBCmd.Args) < 2 {
		return diceerrors.ErrWrongArgumentCount("RENAME")
	}

	key := diceDBCmd.Args[0]
	sid, rc := h.shardManager.GetShardInfo(key)

	preCmd := cmd.DiceDBCmd{
		Cmd:  "RENAME",
		Args: []string{key},
	}

	rc <- &ops.StoreOp{
		SeqID:         0,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmd,
		CmdHandlerID:  h.id,
		ShardID:       sid,
		Client:        nil,
		PreProcessing: true,
	}

	return nil
}

// preProcessCopy prepares the COPY command for preprocessing by sending a GET command
// to retrieve the value of the original key. The retrieved value is used later in the
// decomposeCopy function to copy the value to the destination key.
func customProcessCopy(h *BaseCommandHandler, diceDBCmd *cmd.DiceDBCmd) error {
	if len(diceDBCmd.Args) < 2 {
		return diceerrors.ErrWrongArgumentCount("COPY")
	}

	sid, rc := h.shardManager.GetShardInfo(diceDBCmd.Args[0])

	preCmd := cmd.DiceDBCmd{
		Cmd:  "COPY",
		Args: []string{diceDBCmd.Args[0]},
	}

	// Need to get response from both keys to handle Replace or not
	rc <- &ops.StoreOp{
		SeqID:         0,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmd,
		CmdHandlerID:  h.id,
		ShardID:       sid,
		Client:        nil,
		PreProcessing: true,
	}

	return nil
}

// preProcessPFMerge prepares the PFMERGE command for preprocessing by sending GETOBJECT commands
// to retrieve the value of all the keys to be merged with. The retrieved value is used later in the
// decomposePFMERGE function to merge the hll keys to the new key.
func preProcessPFMerge(h *BaseCommandHandler, diceDBCmd *cmd.DiceDBCmd) error {
	if len(diceDBCmd.Args) < 1 {
		return diceerrors.ErrWrongArgumentCount("PFMERGE")
	}

	// Sending GETOBJECT commands for the keys to be merged, which would be used in
	// later stages to merge with destination key in PFMERGE command.
	for i := 1; i < len(diceDBCmd.Args); i++ {
		sid, rc := h.shardManager.GetShardInfo(diceDBCmd.Args[i])
		preCmd := cmd.DiceDBCmd{
			Cmd:  "GETOBJECT",
			Args: []string{diceDBCmd.Args[i]},
		}

		rc <- &ops.StoreOp{
			SeqID:         0,
			RequestID:     GenerateUniqueRequestID(),
			Cmd:           &preCmd,
			CmdHandlerID:  h.id,
			ShardID:       sid,
			Client:        nil,
			PreProcessing: true,
		}
	}

	return nil
}

// preProcessGeoRadiusByMember is a preprocess function for the GEORADIUSBYMEMBER command.
// This command has a STORE/STOREDIST option that allows the user to directly store the result
// of the command with a new key. Hence, if one of these options is given, the command becomes
// a multi-shard command and we need to break it up.
func preProcessGeoRadiusByMember(h *BaseCommandHandler, cd *cmd.DiceDBCmd) error {
	if len(cd.Args) < 4 {
		return diceerrors.ErrWrongArgumentCount("GEORADIUSBYMEMBER")
	}

	ok, _, parseErr := parseGeoRadiusStoreOpts(cd.Args[4:])
	if parseErr != nil {
		return parseErr
	}

	// if STORE/STOREDIST options are not given, we're not doing any pre-processing
	if !ok {
		return nil
	}

	sid, rc := h.shardManager.GetShardInfo(cd.Args[0])

	preCmd := cmd.DiceDBCmd{
		Cmd:  "GEORADIUSBYMEMBER",
		Args: cd.Args,
	}

	rc <- &ops.StoreOp{
		SeqID:         0,
		RequestID:     GenerateUniqueRequestID(),
		Cmd:           &preCmd,
		CmdHandlerID:  h.id,
		ShardID:       sid,
		Client:        nil,
		PreProcessing: true,
	}

	return nil
}
