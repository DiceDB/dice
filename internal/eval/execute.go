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

package eval

import (
	"strings"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

type Eval struct {
	cmd                   *cmd.DiceDBCmd
	client                *comm.Client
	store                 *dstore.Store
	isHTTPOperation       bool
	isWebSocketOperation  bool
	isPreprocessOperation bool
}

func NewEval(c *cmd.DiceDBCmd, client *comm.Client, store *dstore.Store, httpOp, websocketOp, preProcessing bool) *Eval {
	return &Eval{
		cmd:                   c,
		client:                client,
		store:                 store,
		isHTTPOperation:       httpOp,
		isWebSocketOperation:  websocketOp,
		isPreprocessOperation: preProcessing,
	}
}

func (e *Eval) PreProcessCommand() *EvalResponse {
	if f, ok := PreProcessing[e.cmd.Cmd]; ok {
		return f(e.cmd.Args, e.store)
	}
	return &EvalResponse{Result: nil, Error: diceerrors.ErrInternalServer}
}

func (e *Eval) ExecuteCommand() *EvalResponse {
	diceCmd, ok := DiceCmds[e.cmd.Cmd]
	if !ok {
		return &EvalResponse{Result: diceerrors.NewErrWithFormattedMessage("unknown command '%s', with args beginning with: %s", e.cmd.Cmd, strings.Join(e.cmd.Args, " ")), Error: nil}
	}

	// Temporary logic till we move all commands to new eval logic.
	// MigratedDiceCmds map contains refactored eval commands
	// For any command we will first check in the existing map
	// if command is NA then we will check in the new map
	// Check if the dice command has been migrated
	if diceCmd.IsMigrated {
		// ===============================================================================
		// dealing with store object is not recommended for all commands
		// These operations are specialised for the commands which requires
		// transferring data across multiple shards. e.g. COPY, RENAME, PFMERGE
		// ===============================================================================
		if e.cmd.InternalObjs != nil {
			// This involves handling object at store level, evaluating it, modifying it, and then storing it back.
			return diceCmd.StoreObjectEval(e.cmd, e.store)
		}

		// If the 'Obj' field is nil, handle the command using the arguments.
		// This path likely involves evaluating the command based on its provided arguments.
		return diceCmd.NewEval(e.cmd.Args, e.store)
	}

	// The following commands could be handled at the shard level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	switch diceCmd.Name {
	// Old implementation kept as it is, but we will be moving
	// to the new implementation soon for all commands
	case auth.Cmd:
		return &EvalResponse{Result: EvalAUTH(e.cmd.Args, e.client), Error: nil}
	case "ABORT":
		return &EvalResponse{Result: clientio.RespOK, Error: nil}
	default:
		return &EvalResponse{Result: diceCmd.Eval(e.cmd.Args, e.store), Error: nil}
	}
}
