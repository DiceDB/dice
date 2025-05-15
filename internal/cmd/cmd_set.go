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

const (
	MaxEXDurationSec = 365 * 24 * 60 * 60 // 1 year in seconds
)

// TODO: Make the SET command return 1 if the SET operation was successful and 0 if the key was not set or updated.
// This should involve checking of the old value and the new value.
var cSET = &CommandMeta{
	Name:      "SET",
	Syntax:    "SET key value [EX seconds | PX milliseconds] [EXAT timestamp | PXAT timestamp] [XX | NX] [KEEPTTL]",
	HelpShort: "SET puts or updates an existing value for a key",
	HelpLong: `
SET puts or updates an existing value for a key.

SET stores the value as its native type - be it int or string. SET supports the following options:

- EX seconds: set the expiration time in seconds
- PX milliseconds: set the expiration time in milliseconds
- EXAT timestamp: set the expiration time in seconds since epoch
- PXAT timestamp: set the expiration time in milliseconds since epoch
- XX: only set the key if it already exists
- NX: only set the key if it does not already exist
- KEEPTTL: keep the existing TTL of the key even if some expiration param like EX, etc is provided

Returns "OK" if the SET operation was successful.
	`,
	Examples: `
localhost:7379> SET k 43
OK
localhost:7379> SET k 43 EX 10
OK
localhost:7379> SET k 43 PX 10000
OK
localhost:7379> SET k 43 EXAT 1772377267
OK
localhost:7379> SET k 43 PXAT 1772377267000
OK
localhost:7379> SET k 43 XX
OK
localhost:7379> SET k 43 NX
OK
localhost:7379> SET k 43 KEEPTTL
OK
	`,
	Eval:    evalSET,
	Execute: executeSET,
}

func init() {
	CommandRegistry.AddCommand(cSET)
}

func newSETRes() *CmdRes {
	return &CmdRes{
		Rs: &wire.Result{
			Message:  "OK",
			Status:   wire.Status_OK,
			Response: &wire.Result_SETRes{SETRes: &wire.SETRes{}},
		},
	}
}

var (
	SETResNilRes = newSETRes()
	SETResOKRes  = newSETRes()
)

// parseParams parses the parameters for the any command
// and returns a map of the parameters and the remainder of the arguments
// as non-params.
func parseParams(args []string) (params map[types.Param]string, nonParams []string) {
	params = map[types.Param]string{}
	for i := 0; i < len(args); i++ {
		arg := types.Param(strings.ToUpper(args[i]))
		switch arg {
		case types.EX, types.PX, types.EXAT, types.PXAT:
			params[arg] = args[i+1]
			i++
		case types.XX, types.NX, types.KEEPTTL, types.LT, types.GT, types.CH, types.INCR:
			params[arg] = "true"
		default:
			nonParams = append(nonParams, args[i])
		}
	}
	return params, nonParams
}

//nolint:gocyclo
func evalSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return SETResNilRes, errors.ErrWrongArgumentCount("SET")
	}

	var key, value = c.C.Args[0], c.C.Args[1]

	params, nonParams := parseParams(c.C.Args[2:])
	if len(nonParams) > 0 {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	}

	// Raise errors if incompatible parameters are provided
	// in one command
	if params[types.EX] != "" && params[types.PX] != "" {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	} else if params[types.EX] != "" && params[types.EXAT] != "" {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	} else if params[types.EX] != "" && params[types.PXAT] != "" {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	} else if params[types.PX] != "" && params[types.EXAT] != "" {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	} else if params[types.PX] != "" && params[types.PXAT] != "" {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	} else if params[types.EXAT] != "" && params[types.PXAT] != "" {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	} else if params[types.XX] != "" && params[types.NX] != "" {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	} else if params[types.KEEPTTL] != "" && (params[types.EX] != "" || params[types.PX] != "" || params[types.EXAT] != "" || params[types.PXAT] != "") {
		return SETResNilRes, errors.ErrInvalidSyntax("SET")
	}

	var err error
	var exDurationSec, exDurationMs int64

	// Default to -1 to indicate that the value is not set
	// and the key will not expire
	exDurationMs = -1

	if params[types.EX] != "" {
		exDurationSec, err = strconv.ParseInt(params[types.EX], 10, 64)
		if err != nil {
			return SETResNilRes, errors.ErrInvalidValue("SET", "EX")
		}
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return SETResNilRes, errors.ErrInvalidValue("SET", "EX")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[types.PX] != "" {
		exDurationMs, err = strconv.ParseInt(params[types.PX], 10, 64)
		if err != nil {
			return SETResNilRes, errors.ErrInvalidValue("SET", "PX")
		}
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return SETResNilRes, errors.ErrInvalidValue("SET", "PX")
		}
	}

	if params[types.EXAT] != "" {
		tv, err := strconv.ParseInt(params[types.EXAT], 10, 64)
		if err != nil {
			return SETResNilRes, errors.ErrInvalidValue("SET", "EXAT")
		}
		exDurationSec = tv - time.Now().Unix()
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return SETResNilRes, errors.ErrInvalidValue("SET", "EXAT")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[types.PXAT] != "" {
		tv, err := strconv.ParseInt(params[types.PXAT], 10, 64)
		if err != nil {
			return SETResNilRes, errors.ErrInvalidValue("SET", "PXAT")
		}
		exDurationMs = tv - time.Now().UnixMilli()
		if exDurationMs <= 0 || exDurationMs >= (MaxEXDurationSec*1000) {
			return SETResNilRes, errors.ErrInvalidValue("SET", "PXAT")
		}
	}

	existingObj := s.Get(key)

	// TODO: Add check for the type before doing the operation
	// The scope of this is not clear and hence need some thought
	// on how the database should react when a SET is called on
	// a key that is not of the type that is being set.
	// Example: existing key is of type string
	// and now set is called with different type.
	// Or, SET is called on a key that has some other type, like HLL
	// Should we reject the operation?
	// Or, should we convert the existing type to the new type?
	// Or, should we just overwrite the value?

	// If XX is provided and the key does not exist, return nil
	// XX: only set the key if it already exists
	// So, if it does not exist, we return nil and move on
	if params[types.XX] != "" && existingObj == nil {
		return SETResOKRes, nil
	}

	// If NX is provided and the key already exists, return nil
	// NX: only set the key if it does not already exist
	// So, if it does exist, we return nil and move on
	if params[types.NX] != "" && existingObj != nil {
		return SETResOKRes, nil
	}

	newObj := CreateObjectFromValue(s, value, exDurationMs)
	s.Put(key, newObj, dstore.WithKeepTTL(params[types.KEEPTTL] != ""))

	return SETResOKRes, nil
}

func executeSET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return SETResNilRes, errors.ErrWrongArgumentCount("SET")
	}
	shard := sm.GetShardForKey(c.C.Args[0])
	return evalSET(c, shard.Thread.Store())
}

func CreateObjectFromValue(s *dstore.Store, value string, expiryMs int64) *object.Obj {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return s.NewObj(intValue, expiryMs, object.ObjTypeInt)
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err == nil {
		return s.NewObj(floatValue, expiryMs, object.ObjTypeFloat)
	}
	return s.NewObj(value, expiryMs, object.ObjTypeString)
}
