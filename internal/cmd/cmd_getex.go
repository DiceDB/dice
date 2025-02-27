// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

const PERSIST = "PERSIST"

var cGETEX = &CommandMeta{
	Name:      "GETEX",
	HelpShort: "Get the value of key and optionally set its expiration.",
	Eval:      evalGETEX,
	Execute:   executeGETEX,
}

func init() {
	CommandRegistry.AddCommand(cGETEX)
}

func evalGETEX(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("GETEX")
	}

	var key = c.C.Args[0]
	params := map[string]string{}
	for i := 1; i < len(c.C.Args); i++ {
		arg := strings.ToUpper(c.C.Args[i])
		switch arg {
		case EX, PX, EXAT, PXAT:
			params[arg] = c.C.Args[i+1]
			i++
		case PERSIST:
			params[arg] = "true"
		}
	}

	// Raise errors if incompatible parameters are provided
	// in one command
	if params[EX] != "" && params[PX] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("GETEX")
	} else if params[EX] != "" && params[EXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("GETEX")
	} else if params[EX] != "" && params[PXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("GETEX")
	} else if params[PX] != "" && params[EXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("GETEX")
	} else if params[PX] != "" && params[PXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("GETEX")
	} else if params[EXAT] != "" && params[PXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("GETEX")
	} else if params[PERSIST] != "" && (params[EX] != "" || params[PX] != "" || params[EXAT] != "" || params[PXAT] != "") {
		return cmdResNil, errors.ErrInvalidSyntax("GETEX")
	}
	var err error
	var exDurationSec, exDurationMs int64

	fmt.Println(params)

	// Default to -1 to indicate that the value is not set
	// and the key will not expire
	exDurationMs = -1

	if params[EX] != "" {
		exDurationSec, err = strconv.ParseInt(params[EX], 10, 64)
		if err != nil {
			return cmdResNil, errors.ErrInvalidValue("GETEX", "EX")
		}
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return cmdResNil, errors.ErrInvalidValue("GETEX", "EX")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[PX] != "" {
		exDurationMs, err = strconv.ParseInt(params[PX], 10, 64)
		if err != nil {
			return cmdResNil, errors.ErrInvalidValue("GETEX", "PX")
		}
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return cmdResNil, errors.ErrInvalidValue("GETEX", "PX")
		}
	}

	if params[EXAT] != "" {
		tv, err := strconv.ParseInt(params[EXAT], 10, 64)
		if err != nil {
			return cmdResNil, errors.ErrInvalidValue("GETEX", "EXAT")
		}
		exDurationSec = tv - utils.GetCurrentTime().Unix()
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return cmdResNil, errors.ErrInvalidValue("GETEX", "EXAT")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[PXAT] != "" {
		tv, err := strconv.ParseInt(params[PXAT], 10, 64)
		if err != nil {
			return cmdResNil, errors.ErrInvalidValue("GETEX", "PXAT")
		}
		exDurationMs = tv - utils.GetCurrentTime().UnixMilli()
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return cmdResNil, errors.ErrInvalidValue("GETEX", "PXAT")
		}
	}

	existingObj := s.Get(key)
	if existingObj == nil {
		return cmdResNil, nil
	}

	resp, err := evalGET(&Cmd{
		C: &wire.Command{
			Cmd:  "GET",
			Args: []string{key},
		}}, s)
	if err != nil {
		return resp, err
	}

	if params[PERSIST] != "" {
		dstore.DelExpiry(existingObj, s)
	} else if exDurationMs != -1 {
		s.SetExpiry(existingObj, exDurationMs)
	}

	return resp, nil
}

func executeGETEX(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) == 0 {
		return cmdResNil, errors.ErrWrongArgumentCount("GETEX")
	}

	shard := sm.GetShardForKey(c.C.Args[0])
	return evalGETEX(c, shard.Thread.Store())
}
