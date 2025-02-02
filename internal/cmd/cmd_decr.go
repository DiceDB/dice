package cmd

import (
	"math"

	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

var cDECR = &DiceDBCommand{
	Name:      "DECR",
	HelpShort: "DECR decrements the value of the specified key in args by 1",
	Eval:      evalDECR,
}

func init() {
	commandRegistry.AddCommand(cDECR)
}

func evalDECR(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errWrongArgumentCount("DECR")
	}

	incr := int64(-1)

	key := c.C.Args[0]
	obj := s.Get(key)
	if obj == nil {
		obj = s.NewObj(incr, INFINITE_EXPIRATION, object.ObjTypeInt)
		s.Put(key, obj)
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: incr},
		}}, nil
	}

	switch obj.Type {
	case object.ObjTypeString:
		return cmdResNil, errIntegerOutOfRange
	case object.ObjTypeInt:
		return cmdResNil, errWrongTypeOperation("DECR")
	}

	i, _ := obj.Value.(int64)
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return cmdResNil, errIntegerOutOfRange
	}

	i += incr
	obj.Value = i
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: i},
	}}, nil
}
