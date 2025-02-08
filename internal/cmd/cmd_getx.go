// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

const PERSIST = "PERSIST"

var cGETX = &DiceDBCommand{
	Name:      "GETEX",
	HelpShort: "Get the value of key and optionally set its expiration.",
	Eval:      evalGETEX,
}

func init() {
	commandRegistry.AddCommand(cGETX)
}

// GETEX key [EX seconds | PX milliseconds | EXAT unix-time-seconds |
// PXAT unix-time-milliseconds | PERSIST]
// Get the value of key and optionally set its expiration.
// GETEX is similar to GET, but is a write command with additional options.
// The GETEX command supports a set of options that modify its behavior:
// EX seconds -- Set the specified expire time, in seconds.
// PX milliseconds -- Set the specified expire time, in milliseconds.
// EXAT timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds.
// PXAT timestamp-milliseconds -- Set the specified Unix time at which the key will expire, in milliseconds.
// PERSIST -- Remove the time to live associated with the key.
// The RESP value of the key is encoded and then returned
// evalGET returns response.RespNIL if key is expired or it does not exist
func evalGETEX(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, errWrongArgumentCount("GETEX")
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
		return cmdResNil, errInvalidSyntax("GETEX")
	} else if params[EX] != "" && params[EXAT] != "" {
		return cmdResNil, errInvalidSyntax("GETEX")
	} else if params[EX] != "" && params[PXAT] != "" {
		return cmdResNil, errInvalidSyntax("GETEX")
	} else if params[PX] != "" && params[EXAT] != "" {
		return cmdResNil, errInvalidSyntax("GETEX")
	} else if params[PX] != "" && params[PXAT] != "" {
		return cmdResNil, errInvalidSyntax("GETEX")
	} else if params[EXAT] != "" && params[PXAT] != "" {
		return cmdResNil, errInvalidSyntax("GETEX")
	} else if params[PERSIST] != "" && (params[EX] != "" || params[PX] != "" || params[EXAT] != "" || params[PXAT] != "") {
		return cmdResNil, errInvalidSyntax("GETEX")
	}
	var err error
	var exDurationSec, exDurationMs int64

	// Default to -1 to indicate that the value is not set
	// and the key will not expire
	exDurationMs = -1

	if params[EX] != "" {
		exDurationSec, err = strconv.ParseInt(params[EX], 10, 64)
		if err != nil {
			return cmdResNil, errInvalidValue("GETEX", "EX")
		}
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return cmdResNil, errInvalidValue("GETEX", "EX")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[PX] != "" {
		exDurationMs, err = strconv.ParseInt(params[PX], 10, 64)
		if err != nil {
			return cmdResNil, errInvalidValue("GETEX", "PX")
		}
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return cmdResNil, errInvalidValue("GETEX", "PX")
		}
	}

	if params[EXAT] != "" {
		tv, err := strconv.ParseInt(params[EXAT], 10, 64)
		if err != nil {
			return cmdResNil, errInvalidValue("GETEX", "EXAT")
		}
		exDurationSec = tv - utils.GetCurrentTime().Unix()
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return cmdResNil, errInvalidValue("GETEX", "EXAT")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[PXAT] != "" {
		tv, err := strconv.ParseInt(params[PXAT], 10, 64)
		if err != nil {
			return cmdResNil, errInvalidValue("GETEX", "PXAT")
		}
		exDurationMs = tv - utils.GetCurrentTime().UnixMilli()
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return cmdResNil, errInvalidValue("GETEX", "PXAT")
		}
	}

	// Get the key from the hash table
	obj := s.Get(key)

	if obj == nil {
		return cmdResNil, nil
	}

	if object.AssertType(obj.Type, object.ObjTypeSet) == nil ||
		object.AssertType(obj.Type, object.ObjTypeJSON) == nil {
		return cmdResNil, errWrongTypeOperation("GETEX")
	}

	// Get EvalResponse with correct data type
	getResp, err := evalGET(&Cmd{
		C: &wire.Command{
			Cmd:  "GET",
			Args: []string{key},
		},
	}, s)

	// If there is an error return the error response
	if err != nil {
		return getResp, err
	}

	if params[PERSIST] != "" {
		dstore.DelExpiry(obj, s)
	} else {
		s.SetExpiry(obj, exDurationMs)
	}

	// return an EvalResponse with the value
	return getResp, nil
}
