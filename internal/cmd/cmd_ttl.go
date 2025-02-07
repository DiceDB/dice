package cmd

import (
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cTTL = &DiceDBCommand{
	Name:      "TTL",
	HelpShort: "TTL return the remaining time to live of a key that has an expiration set",
	Eval:      evalTTL,
}

func init() {
	commandRegistry.AddCommand(cTTL)
}

func evalTTL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("TTL")
	}

	var key = c.C.Args[0]

	obj := s.Get(key)

	if obj == nil {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: -2},
		}}, nil
	}

	exp, isExpirySet := dstore.GetExpiry(obj, s)

	if !isExpirySet {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: -1},
		}}, nil
	}

	durationMs := exp - uint64(utils.GetCurrentTime().UnixMilli())

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: int64(durationMs / 1000)},
	}}, nil

}
