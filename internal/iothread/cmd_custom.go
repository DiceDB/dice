package iothread

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
)

// RespAuth returns with an encoded "OK" if the user is authenticated
// If the user is not authenticated, it returns with an encoded error message
func (t *BaseIOThread) RespAuth(args []string) interface{} {
	// Check for incorrect number of arguments (arity error).
	if len(args) < 1 || len(args) > 2 {
		return diceerrors.ErrWrongArgumentCount("AUTH")
	}

	if config.DiceConfig.Auth.Password == "" {
		return diceerrors.ErrAuth
	}

	username := config.DiceConfig.Auth.UserName
	var password string

	if len(args) == 1 {
		password = args[0]
	} else {
		username, password = args[0], args[1]
	}

	if err := t.Session.Validate(username, password); err != nil {
		return err
	}

	return clientio.OK
}

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
