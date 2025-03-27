// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package cmd

import (
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/shardmanager"
	dstore "github.com/dicedb/dice/internal/store"
)

const (
	EX               = "EX"
	PX               = "PX"
	EXAT             = "EXAT"
	PXAT             = "PXAT"
	XX               = "XX"
	NX               = "NX"
	KEEPTTL          = "KEEPTTL"
	GET              = "GET"
	MaxEXDurationSec = 365 * 24 * 60 * 60 // 1 year in seconds
)

var cSET = &CommandMeta{
	Name:      "SET",
	Syntax:    "SET key value [EX seconds] [PX milliseconds] [EXAT timestamp] [PXAT timestamp] [XX] [NX] [KEEPTTL]",
	HelpShort: "SET puts or updates an existing <key, value> pair",
	HelpLong: `
SET puts or updates an existing <key, value> pair.

SET identifies the type of the value based on the value itself. If the value is an integer,
it will be stored as an integer. Otherwise, it will be stored as a string.

- EX seconds: Set the expiration time in seconds
- PX milliseconds: Set the expiration time in milliseconds
- EXAT timestamp: Set the expiration time in seconds since epoch
- PXAT timestamp: Set the expiration time in milliseconds since epoch
- XX: Only set the key if it already exists
- NX: Only set the key if it does not already exist
- KEEPTTL: Keep the existing TTL of the key
- GET: Return the value of the key after setting it

Returns "OK" if the key was set or updated. Returns (nil) if the key was not set or updated.
Returns the value of the key if the GET option is provided.
	`,
	Examples: `
localhost:7379> SET k 43
OK OK
localhost:7379> SET k 43 EX 10
OK OK
localhost:7379> SET k 43 PX 10000
OK OK
localhost:7379> SET k 43 EXAT 1772377267
OK OK
localhost:7379> SET k 43 PXAT 1772377267000
OK OK
localhost:7379> SET k 43 XX
OK OK
localhost:7379> SET k 43 NX
OK (nil)
localhost:7379> SET k 43 KEEPTTL
OK OK
localhost:7379> SET k 43 GET
OK 43
	`,
	Eval:    evalSET,
	Execute: executeSET,
}

func init() {
	CommandRegistry.AddCommand(cSET)
}

//nolint:gocyclo
func evalSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("SET")
	}

	var key, value = c.C.Args[0], c.C.Args[1]
	params := map[string]string{}

	for i := 2; i < len(c.C.Args); i++ {
		arg := strings.ToUpper(c.C.Args[i])
		switch arg {
		case EX, PX, EXAT, PXAT:
			params[arg] = c.C.Args[i+1]
			i++
		case XX, NX, KEEPTTL, "GET":
			params[arg] = "true"
		}
	}

	// Raise errors if incompatible parameters are provided
	// in one command
	if params[EX] != "" && params[PX] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("SET")
	} else if params[EX] != "" && params[EXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("SET")
	} else if params[EX] != "" && params[PXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("SET")
	} else if params[PX] != "" && params[EXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("SET")
	} else if params[PX] != "" && params[PXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("SET")
	} else if params[EXAT] != "" && params[PXAT] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("SET")
	} else if params[XX] != "" && params[NX] != "" {
		return cmdResNil, errors.ErrInvalidSyntax("SET")
	} else if params[KEEPTTL] != "" && (params[EX] != "" || params[PX] != "" || params[EXAT] != "" || params[PXAT] != "") {
		return cmdResNil, errors.ErrInvalidSyntax("SET")
	}

	var err error
	var exDurationSec, exDurationMs int64

	// Default to -1 to indicate that the value is not set
	// and the key will not expire
	exDurationMs = -1

	if params[EX] != "" {
		exDurationSec, err = strconv.ParseInt(params[EX], 10, 64)
		if err != nil {
			return cmdResNil, errors.ErrInvalidValue("SET", "EX")
		}
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return cmdResNil, errors.ErrInvalidValue("SET", "EX")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[PX] != "" {
		exDurationMs, err = strconv.ParseInt(params[PX], 10, 64)
		if err != nil {
			return cmdResNil, errors.ErrInvalidValue("SET", "PX")
		}
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return cmdResNil, errors.ErrInvalidValue("SET", "PX")
		}
	}

	if params[EXAT] != "" {
		tv, err := strconv.ParseInt(params[EXAT], 10, 64)
		if err != nil {
			return cmdResNil, errors.ErrInvalidValue("SET", "EXAT")
		}
		exDurationSec = tv - utils.GetCurrentTime().Unix()
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return cmdResNil, errors.ErrInvalidValue("SET", "EXAT")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[PXAT] != "" {
		tv, err := strconv.ParseInt(params[PXAT], 10, 64)
		if err != nil {
			return cmdResNil, errors.ErrInvalidValue("SET", "PXAT")
		}
		exDurationMs = tv - utils.GetCurrentTime().UnixMilli()
		if exDurationMs <= 0 || exDurationMs >= (MaxEXDurationSec*1000) {
			return cmdResNil, errors.ErrInvalidValue("SET", "PXAT")
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
	if params[XX] != "" && existingObj == nil {
		return cmdResNil, nil
	}

	// If NX is provided and the key already exists, return nil
	// NX: only set the key if it does not already exist
	// So, if it does exist, we return nil and move on
	if params[NX] != "" && existingObj != nil {
		return cmdResNil, nil
	}

	newObj := CreateObjectFromValue(s, value, exDurationMs)
	s.Put(key, newObj, dstore.WithKeepTTL(params[KEEPTTL] != ""))

	if params[GET] != "" {
		// TODO: Optimize this because we have alread fetched the
		// object in the existingObj variable. We can avoid executing
		// the GET command again.

		// If existingObj is nil then the key does not exist
		// and we return nil
		if existingObj == nil {
			return cmdResNil, nil
		}

		return cmdResFromObject(existingObj)
	}

	return cmdResOK, nil
}

func executeSET(c *Cmd, sm *shardmanager.ShardManager) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errors.ErrWrongArgumentCount("SET")
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
	} else {
		return s.NewObj(value, expiryMs, object.ObjTypeString)
	}
}
