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

const (
	EX               = "EX"
	PX               = "PX"
	EXAT             = "EXAT"
	PXAT             = "PXAT"
	XX               = "XX"
	NX               = "NX"
	KEEPTTL          = "KEEPTTL"
	GET              = "GET"
	MaxEXDurationSec = 10 * 24 * 60 * 60 // 10 days in seconds
)

var cSET = &DiceDBCommand{
	Name:      "SET",
	HelpShort: "SET puts a new <key, value> pair. If the key already exists then the value will be overwritten.",
	Eval:      evalSET,
}

func init() {
	commandRegistry.AddCommand(cSET)
}

//nolint:gocyclo
func evalSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errWrongArgumentCount("SET")
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
		return cmdResNil, errInvalidSyntax("SET")
	} else if params[EX] != "" && params[EXAT] != "" {
		return cmdResNil, errInvalidSyntax("SET")
	} else if params[EX] != "" && params[PXAT] != "" {
		return cmdResNil, errInvalidSyntax("SET")
	} else if params[PX] != "" && params[EXAT] != "" {
		return cmdResNil, errInvalidSyntax("SET")
	} else if params[PX] != "" && params[PXAT] != "" {
		return cmdResNil, errInvalidSyntax("SET")
	} else if params[EXAT] != "" && params[PXAT] != "" {
		return cmdResNil, errInvalidSyntax("SET")
	} else if params[XX] != "" && params[NX] != "" {
		return cmdResNil, errInvalidSyntax("SET")
	} else if params[KEEPTTL] != "" && (params[EX] != "" || params[PX] != "" || params[EXAT] != "" || params[PXAT] != "") {
		return cmdResNil, errInvalidSyntax("SET")
	}

	var err error
	var exDurationSec, exDurationMs int64

	// Default to -1 to indicate that the value is not set
	// and the key will not expire
	exDurationMs = -1

	if params[EX] != "" {
		exDurationSec, err = strconv.ParseInt(params[EX], 10, 64)
		if err != nil {
			return cmdResNil, errInvalidValue("SET", "EX")
		}
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return cmdResNil, errInvalidValue("SET", "EX")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[PX] != "" {
		exDurationMs, err = strconv.ParseInt(params[PX], 10, 64)
		if err != nil {
			return cmdResNil, errInvalidValue("SET", "PX")
		}
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return cmdResNil, errInvalidValue("SET", "PX")
		}
	}

	if params[EXAT] != "" {
		tv, err := strconv.ParseInt(params[EXAT], 10, 64)
		if err != nil {
			return cmdResNil, errInvalidValue("SET", "EXAT")
		}
		exDurationSec = tv - utils.GetCurrentTime().Unix()
		if exDurationSec <= 0 || exDurationSec >= MaxEXDurationSec {
			return cmdResNil, errInvalidValue("SET", "EXAT")
		}
		exDurationMs = exDurationSec * 1000
	}

	if params[PXAT] != "" {
		tv, err := strconv.ParseInt(params[PXAT], 10, 64)
		if err != nil {
			return cmdResNil, errInvalidValue("SET", "PXAT")
		}
		exDurationMs = tv - utils.GetCurrentTime().UnixMilli()
		if exDurationMs <= 0 || exDurationMs >= MaxEXDurationSec {
			return cmdResNil, errInvalidValue("SET", "PXAT")
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

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		s.Put(key, s.NewObj(intValue, exDurationMs, object.ObjTypeInt), dstore.WithKeepTTL(params[KEEPTTL] != ""))
	} else {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			s.Put(key, s.NewObj(floatValue, exDurationMs, object.ObjTypeFloat), dstore.WithKeepTTL(params[KEEPTTL] != ""))
		} else {
			s.Put(key, s.NewObj(value, exDurationMs, object.ObjTypeString), dstore.WithKeepTTL(params[KEEPTTL] != ""))
		}
	}

	if params[GET] != "" {
		// TODO: Optimize this because we have alread fetched the
		// object in the existingObj variable. We can avoid executing
		// the GET command again.

		// If existingObj is nil then the key does not exist
		// and we return nil
		if existingObj == nil {
			return cmdResNil, nil
		}

		// If existingObj is not nil then the key exists
		// and we need to fetch the value of the key
		crExistingKey, err := evalGET(&Cmd{
			C: &wire.Command{
				Cmd:  "GET",
				Args: []string{key},
			},
		}, s)
		if err != nil {
			return crExistingKey, err
		}
	}

	return cmdResOK, nil
}
