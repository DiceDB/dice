// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package commandhandler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
)

// RespPING evaluates the PING command and returns the appropriate response.
// If no arguments are provided, it responds with "PONG" (standard behavior).
// If an argument is provided, it returns the argument as the response.
// If more than one argument is provided, it returns an arity error.
func RespPING(args []string) interface{} {
	// Check for incorrect number of arguments (arity error).
	if len(args) >= 2 {
		return diceerrors.ErrWrongArgumentCount("PING") // Return an error if more than one argument is provided.
	}

	// If no arguments are provided, return the standard "PONG" response.
	if len(args) == 0 {
		return "PONG"
	}
	// If one argument is provided, return the argument as the response.
	return args[0]
}

func RespEcho(args []string) interface{} {
	if len(args) != 1 {
		return diceerrors.ErrWrongArgumentCount("ECHO")
	}

	return args[0]
}

func RespHello(args []string) interface{} {
	if len(args) > 1 {
		return diceerrors.NewErrArity("HELLO")
	}

	var resp []interface{}
	serverID := fmt.Sprintf("%s:%d", config.DiceConfig.RespServer.Addr, config.DiceConfig.RespServer.Port)
	resp = append(resp,
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{})

	return resp
}

func RespSleep(args []string) interface{} {
	if len(args) != 1 {
		return diceerrors.NewErrArity("SLEEP")
	}

	durationSec, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
	}
	time.Sleep(time.Duration(durationSec) * time.Second)
	return clientio.OK
}
