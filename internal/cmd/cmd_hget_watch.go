// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"google.golang.org/protobuf/types/known/structpb"
)

var cHGETWATCH = &CommandMeta{
	Name:      "HGET.WATCH",
	Syntax:    "HGET.WATCH key field",
	HelpShort: "HGET.WATCH creates a query subscription over the HGET command",
	HelpLong: `
HGET.WATCH creates a query subscription over the HGET command. The client invoking the command
will receive the output of the HGET command (not just the notification) whenever the value against
the key and field is updated.

You can update the key in any other client. The HGET.WATCH client will receive the updated value.
	`,
	Examples: `
client1:7379>  HSET k1 f1 v1
OK 1
client1:7379> HGET.WATCH k1 f1
entered the watch mode for HGET.WATCH k1 f1


client2:7379>  HSET k1 f1 v2
OK 0


client1:7379> ...
entered the watch mode for HGET.WATCH k1 f1
OK [fingerprint=3432795955] v2
	`,
	Eval:    evalHGETWATCH,
	Execute: executeHGETWATCH,
}

func init() {
	CommandRegistry.AddCommand(cHGETWATCH)
}

func evalHGETWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	r, err := evalHGET(c, s)
	if err != nil {
		return nil, err
	}

	if r.R.Attrs == nil {
		r.R.Attrs = &structpb.Struct{
			Fields: make(map[string]*structpb.Value),
		}
	}

	r.R.Attrs.Fields["fingerprint"] = structpb.NewStringValue(strconv.FormatUint(uint64(c.Fingerprint()), 10))
	return r, nil
}

func executeHGETWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errors.ErrWrongArgumentCount("HGET.WATCH")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalHGETWATCH(c, shard.Thread.Store())
}
