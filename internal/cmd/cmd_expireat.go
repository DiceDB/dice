package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
	"strconv"
)

var cEXPIREAT = &DiceDBCommand{
	Name:      "EXPIREAT",
	HelpShort: "EXPIREAT sets the expiration time of a key as an absolute Unix timestamp (in seconds)",
	Eval:      evalEXPIREAT,
}

func init() {
	commandRegistry.AddCommand(cEXPIREAT)
}

func evalEXPIREAT(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errWrongArgumentCount("EXPIREAT")
	}

	var key = c.C.Args[0]
	exUnixTimeSec, err := strconv.ParseInt(c.C.Args[1], 10, 64)

	if err != nil {
		return cmdResNil, errInvalidExpireTime("EXPIRE")
	}

	if exUnixTimeSec < 0 {
		return cmdResNil, errInvalidExpireTime("EXPIRE")
	}

	isExpirySet, err2 := dstore.EvaluateAndSetExpiry(c.C.Args[2:], exUnixTimeSec, key, s)

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
