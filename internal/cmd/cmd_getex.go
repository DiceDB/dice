// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/internal/types"
	"github.com/dicedb/dicedb-go/wire"
)

const PERSIST = "PERSIST"

var cGETEX = &CommandMeta{
	Name:      "GETEX",
	Syntax:    "GETEX key [EX seconds | PX milliseconds] [EXAT timestamp-seconds | PXAT timestamp-milliseconds] [PERSIST]",
	HelpShort: "GETEX gets the value of key and optionally set its expiration.",
	HelpLong: `
GETEX gets the value of key and optionally set its expiration. The behavior of the command
is similar to the GET command with the addition of the ability to set an expiration on the key.

The command returns (nil) if the key does not exist. The command supports the following options:

- EX seconds: Set the expiration to seconds from now.
- PX milliseconds: Set the expiration to milliseconds from now.
- EXAT timestamp: Set the expiration to a Unix timestamp.
- PXAT timestamp: Set the expiration to a Unix timestamp in milliseconds.
- PERSIST: Remove the expiration from the key.
	`,
	Examples: `
localhost:7379> SET k v
OK 
localhost:7379> GETEX k EX 1000
OK "v"
localhost:7379> TTL k
OK 996
localhost:7379> GETEX k PX 200000
OK "v"
localhost:7379> GETEX k EXAT 1772377267
OK "v"
localhost:7379> GETEX k PXAT 1772377267000
OK "v"
localhost:7379> GETEX k PERSIST
OK "v"
localhost:7379> EXPIRETIME k
OK -1
	`,
	Eval:    evalGETEX,
	Execute: executeGETEX,
}

func init() {
	CommandRegistry.AddCommand(cGETEX)
}

func newGETEXRes(obj *object.Obj) *CmdRes {
	value, err := getWireValueFromObj(obj)
	if err != nil {
		return &CmdRes{
			Rs: &wire.Result{
				Message: err.Error(),
				Status:  wire.Status_ERR,
			},
		}
	}
	return &CmdRes{
		Rs: &wire.Result{
			Message: "OK",
			Status:  wire.Status_OK,
			Response: &wire.Result_GETEXRes{
				GETEXRes: &wire.GETEXRes{
					Value: value,
				},
			},
		},
	}
}

var (
	GETEXResNilRes = newGETEXRes(nil)
)

func evalGETEX(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return GETEXResNilRes, errors.ErrWrongArgumentCount("GETEX")
	}

	var key = c.C.Args[0]
	params := map[types.Param]string{}
	for i := 1; i < len(c.C.Args); i++ {
		arg := types.Param(strings.ToUpper(c.C.Args[i]))
		switch arg {
		case types.EX, types.PX, types.EXAT, types.PXAT:
			params[arg] = c.C.Args[i+1]
			i++
		case types.PERSIST:
			params[arg] = "true"
		}
	}

	// Raise errors if incompatible parameters are provided
	// in one command
	if params[types.EX] != "" && params[types.PX] != "" {
		return GETEXResNilRes, errors.ErrInvalidSyntax("GETEX")
	} else if params[types.EX] != "" && params[types.EXAT] != "" {
		return GETEXResNilRes, errors.ErrInvalidSyntax("GETEX")
	} else if params[types.EX] != "" && params[types.PXAT] != "" {
		return GETEXResNilRes, errors.ErrInvalidSyntax("GETEX")
	} else if params[types.PX] != "" && params[types.EXAT] != "" {
		return GETEXResNilRes, errors.ErrInvalidSyntax("GETEX")
	} else if params[types.PX] != "" && params[types.PXAT] != "" {
		return GETEXResNilRes, errors.ErrInvalidSyntax("GETEX")
	} else if params[types.EXAT] != "" && params[types.PXAT] != "" {
		return GETEXResNilRes, errors.ErrInvalidSyntax("GETEX")
	} else if params[types.PERSIST] != "" && (params[types.EX] != "" || params[types.PX] != "" || params[types.EXAT] != "" || params[types.PXAT] != "") {
		return GETEXResNilRes, errors.ErrInvalidSyntax("GETEX")
	}
	var err error
	var exDurationSec, exDurationMs int64

	// Default to -1 to indicate that the value is not set
	// and the key will not expire
	exDurationMs = -1

	if params[types.EX] != "" {
		exDurationSec, err = strconv.ParseInt(params[types.EX], 10, 64)
		if err != nil {
			return GETEXResNilRes, errors.ErrInvalidValue("GETEX", "EX")
		}
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return GETEXResNilRes, errors.ErrInvalidValue("GETEX", "EX")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[types.PX] != "" {
		exDurationMs, err = strconv.ParseInt(params[types.PX], 10, 64)
		if err != nil {
			return GETEXResNilRes, errors.ErrInvalidValue("GETEX", "PX")
		}
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return GETEXResNilRes, errors.ErrInvalidValue("GETEX", "PX")
		}
	}

	if params[types.EXAT] != "" {
		tv, err := strconv.ParseInt(params[types.EXAT], 10, 64)
		if err != nil {
			return GETEXResNilRes, errors.ErrInvalidValue("GETEX", "EXAT")
		}
		exDurationSec = tv - time.Now().Unix()
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return GETEXResNilRes, errors.ErrInvalidValue("GETEX", "EXAT")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[types.PXAT] != "" {
		tv, err := strconv.ParseInt(params[types.PXAT], 10, 64)
		if err != nil {
			return GETEXResNilRes, errors.ErrInvalidValue("GETEX", "PXAT")
		}
		exDurationMs = tv - time.Now().UnixMilli()
		if exDurationMs <= 0 || exDurationMs >= (MaxEXDurationSec*1000) {
			return GETEXResNilRes, errors.ErrInvalidValue("GETEX", "PXAT")
		}
	}

	existingObj := s.Get(key)
	if existingObj == nil {
		return GETEXResNilRes, nil
	}

	if params[types.PERSIST] != "" {
		dstore.DelExpiry(existingObj, s)
	} else if exDurationMs != -1 {
		s.SetExpiry(existingObj, exDurationMs)
	}

	return newGETEXRes(existingObj), nil
}

func executeGETEX(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return GETEXResNilRes, errors.ErrWrongArgumentCount("GETEX")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGETEX(c, shard.Thread.Store())
}
