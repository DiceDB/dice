// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/comm"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

type exDurationState int

const (
	Uninitialized exDurationState = iota
	Initialized
)

var (
	TxnCommands       map[string]bool
	serverID          string
	diceCommandsCount int
)

// EvalResponse represents the response of an evaluation operation for a command from store.
// It contains the sequence ID, the result of the store operation, and any error encountered during the operation.
type EvalResponse struct {
	Result interface{} // Result holds the outcome of the Store operation. Currently, it is expected to be of type []byte, but this may change in the future.
	Error  error       // Error holds any error that occurred during the operation. If no error, it will be nil.
}

// Following functions should be used to create a new EvalResponse with the given result and error.
// These ensure that result and error are mutually exclusive.
// If result is nil, then error should be non-nil and vice versa.

// makeEvalResult creates a new EvalResponse with the given result and nil error.
// This is a helper function to create a new EvalResponse with the given result and nil error.
/**
 * @param {interface{}} result - The result of the store operation.
 * @returns {EvalResponse} A new EvalResponse with the given result and nil error.
 */
func makeEvalResult(result interface{}) *EvalResponse {
	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// makeEvalError creates a new EvalResponse with the given error and nil result.
// This is a helper function to create a new EvalResponse with the given error and nil result.
/**
 * @param {error} err - The error that occurred during the store operation.
 * @returns {EvalResponse} A new EvalResponse with the given error and nil result.
 */
func makeEvalError(err error) *EvalResponse {
	return &EvalResponse{
		Result: nil,
		Error:  err,
	}
}

type jsonOperation string

const (
	IncrBy = "INCRBY"
	MultBy = "MULTBY"
)

const (
	defaultRootPath = "$"
	maxExDuration   = 9223372036854775
	CountConst      = "COUNT"
)

func init() {
	diceCommandsCount = len(DiceCmds)
	TxnCommands = map[string]bool{"EXEC": true, "DISCARD": true}
}

// EvalAUTH returns with an encoded "OK" if the user is authenticated
// If the user is not authenticated, it returns with an encoded error message
// TODO: Needs to be removed after http and websocket migrated to the multithreading
func EvalAUTH(args []string, c *comm.Client) []byte {
	var err error

	if config.Config.Password == "" {
		return diceerrors.NewErrWithMessage("AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?")
	}

	username := config.Config.Username
	var password string

	if len(args) == 1 {
		password = args[0]
	} else if len(args) == 2 {
		username, password = args[0], args[1]
	} else {
		return diceerrors.NewErrArity("AUTH")
	}

	if err = c.Session.Validate(username, password); err != nil {
		return Encode(err, false)
	}
	return RespOK
}

func evalHELLO(args []string, store *dstore.Store) []byte {
	if len(args) > 1 {
		return diceerrors.NewErrArity("HELLO")
	}

	var resp []interface{}
	serverID = fmt.Sprintf("%s:%d", config.Config.Host, config.Config.Port)
	resp = append(resp,
		"proto", 2,
		"id", serverID,
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{})

	return Encode(resp, false)
}

// evalSLEEP sets db to sleep for the specified number of seconds.
// The sleep time should be the only param in args.
// Returns error response if the time param in args is not of integer format.
// evalSLEEP returns response.RespOK after sleeping for mentioned seconds
func evalSLEEP(args []string, store *dstore.Store) []byte {
	if len(args) != 1 {
		return diceerrors.NewErrArity("SLEEP")
	}

	durationSec, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return diceerrors.NewErrWithMessage(diceerrors.IntOrOutOfRangeErr)
	}
	time.Sleep(time.Duration(durationSec) * time.Second)
	return RespOK
}
