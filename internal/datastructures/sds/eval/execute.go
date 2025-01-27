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

	"github.com/dicedb/dice/internal/cmd"
	ds "github.com/dicedb/dice/internal/datastructures"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

type Eval struct {
	op                    *cmd.DiceDBCmd
	store                 *dstore.Store
	isHTTPOperation       bool
	isWebSocketOperation  bool
	isPreprocessOperation bool
}

func NewEval(c *cmd.DiceDBCmd, store *dstore.Store, httpOp, websocketOp, preProcessing bool) *Eval {
	return &Eval{
		op:                    c,
		store:                 store,
		isHTTPOperation:       httpOp,
		isWebSocketOperation:  websocketOp,
		isPreprocessOperation: preProcessing,
	}
}

func (e *Eval) PreProcessCommand() *ds.EvalResponse {
	if f, ok := PreProcessing[e.op.Cmd]; ok {
		return f(e.op.Args, e.store)
	}
	return &ds.EvalResponse{Result: nil, Error: diceerrors.ErrInternalServer}
}

func (e *Eval) ExecuteCommand() *ds.EvalResponse {
	diceCmd, ok := DiceCmds[e.op.Cmd]
	if !ok {
		return &ds.EvalResponse{Result: diceerrors.NewErrWithFormattedMessage("unknown command '%s', with args beginning with: %s", e.op.Cmd, strings.Join(e.op.Args, " ")), Error: nil}
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
		if e.op.InternalObjs != nil {
			// This involves handling object at store level, evaluating it, modifying it, and then storing it back.
			return diceCmd.StoreObjectEval(e.op, e.store)
		}

		// If the 'Obj' field is nil, handle the command using the arguments.
		// This path likely involves evaluating the command based on its provided arguments.
		return diceCmd.NewEval(e.op.Args, e.store)
	}

	// The following commands could be handled at the shard level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	return e.Evaluate()
}
