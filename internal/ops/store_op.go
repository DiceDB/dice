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

package ops

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/eval"
)

type StoreOp struct {
	SeqID         uint8          // SeqID is the sequence id of the operation within a single request (optional, may be used for ordering)
	RequestID     uint32         // RequestID identifies the request that this StoreOp belongs to
	Cmd           *cmd.DiceDBCmd // Cmd is the atomic Store command (e.g., GET, SET)
	ShardID       uint8          // ShardID of the shard on which the Store command will be executed
	CmdHandlerID  string         // CmdHandlerID is the ID of the command handler that sent this Store operation
	Client        *comm.Client   // Client that sent this Store operation. TODO: This can potentially replace the CmdHandlerID in the future
	HTTPOp        bool           // HTTPOp is true if this Store operation is an HTTP operation
	WebsocketOp   bool           // WebsocketOp is true if this Store operation is a Websocket operation
	PreProcessing bool           // PreProcessing indicates whether a comamnd operation requires preprocessing before execution. This is mainly used is multi-step-multi-shard commands
}

// StoreResponse represents the response of a Store operation.
type StoreResponse struct {
	RequestID    uint32             // RequestID that this StoreResponse belongs to
	EvalResponse *eval.EvalResponse // Result of the Store operation, for now the type is set to []byte, but this can change in the future.
	SeqID        uint8              // Sequence ID to maintain the order of responses, used to track the sequence in which operations are processed or received.
}
