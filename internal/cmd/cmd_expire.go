// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"

	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cEXPIRE = &CommandMeta{
	Name:      "EXPIRE",
	HelpShort: "EXPIRE sets an expiry(in seconds) on a specified key",
	Eval:      evalEXPIRE,
}

func init() {
	CommandRegistry.AddCommand(cEXPIRE)
}

func evalEXPIRE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errWrongArgumentCount("EXPIRE")
	}

	var key = c.C.Args[0]
	exDurationSec, err := strconv.ParseInt(c.C.Args[1], 10, 64)

	if err != nil || exDurationSec < 0 {
		return cmdResNil, errInvalidExpireTime("EXPIRE")
	}

	obj := s.Get(key)

	if obj == nil {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: 0},
		}}, nil
	}

	isExpirySet, err2 := dstore.EvaluateAndSetExpiry(c.C.Args[2:], utils.AddSecondsToUnixEpoch(exDurationSec), key, s)

	if err2 != nil {
		return cmdResNil, err2
	}

	if isExpirySet {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: 1},
		}}, nil
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: 0},
	}}, nil
}
