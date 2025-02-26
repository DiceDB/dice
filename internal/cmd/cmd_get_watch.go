// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"google.golang.org/protobuf/types/known/structpb"
)

var cGETWATCH = &CommandMeta{
	Name:      "GET.WATCH",
	HelpShort: "GET.WATCH creates a query subscription over the GET command",
	Eval:      evalGETWATCH,
	Execute:   executeGETWATCH,
}

func init() {
	CommandRegistry.AddCommand(cGETWATCH)
}

func evalGETWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("GET.WATCH")
	}

	r, err := evalGET(c, s)
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

func executeGETWATCH(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGETWATCH(c, shard.Thread.Store())
}
