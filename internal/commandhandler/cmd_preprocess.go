// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
