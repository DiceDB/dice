// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cHANDSHAKE = &CommandMeta{
	Name:      "HANDSHAKE",
	Syntax:    "HANDSHAKE client_id execution_mode",
	HelpShort: "HANDSHAKE tells the server the purpose of the connection",
	HelpLong: `
HANDSHAKE is used to tell the DiceDB server the purpose of the connection. It
registers the client_id and execution_mode.

The client_id is a unique identifier for the client. It can be any string, typically
a UUID.

The execution_mode is the mode of the connection, it can be one of the following:

1. "command" - The client will send commands to the server and receive responses.
2. "watch" - The connection in the watch mode will be used to receive the responses of query subscriptions.

If you use DiceDB SDK or CLI then this HANDSHAKE command is automatically sent when the connection is established
or when you establish a subscription.
	`,
	Examples: `
localhost:7379> HANDSHAKE 4c9d0411-6b28-4ee5-b78a-e7e258afa52f command
OK
	`,
	Eval:    evalHANDSHAKE,
	Execute: executeHANDSHAKE,
}

func init() {
	CommandRegistry.AddCommand(cHANDSHAKE)
}

func newHANDSHAKERes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_HANDSHAKERes{
				HANDSHAKERes: &wire.HANDSHAKERes{},
			},
		},
	}
}

var (
	HANDSHAKEResNilRes = newHANDSHAKERes()
	HANDSHAKEResOKRes  = newHANDSHAKERes()
)

func evalHANDSHAKE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return HANDSHAKEResNilRes, errors.ErrWrongArgumentCount("HANDSHAKE")
	}
	c.ClientID = c.C.Args[0]
	c.Mode = c.C.Args[1]
	return HANDSHAKEResOKRes, nil
}

func executeHANDSHAKE(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey("-")
	return evalHANDSHAKE(c, shard.Thread.Store())
}
