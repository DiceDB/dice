package cmd

import (
	"math"
	"strconv"

	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

var cDECRBY = &DiceDBCommand{
	Name:      "DECRBY",
	HelpShort: "DECRBY decrements the value of the specified key in args by the specified decrement",
	Eval:      evalDECRBY,
}


func init() {
	commandRegistry.AddCommand(cDECRBY)
}

func evalDECRBY(c *Cmd, s *dstore.Store)  (*CmdRes, error) {
	if len(c.C.Args) != 2 {
		return cmdResNil, errWrongArgumentCount("DECRBY")
	}

	delta , err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil {
		return cmdResNil, errIntegerOutOfRange
	}

	return incrDecr(c,s,-delta)
}

func incrDecr(c *Cmd, s *dstore.Store, delta int64) (*CmdRes, error) {
	key := c.C.Args[0]
	obj := s.Get(key)
	if obj == nil {
		obj = s.NewObj(delta, -1, object.ObjTypeInt)
		s.Put(key, obj)
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: delta},
		}}, nil
	}

	switch obj.Type {
	case object.ObjTypeInt:
		break
	default:
		return cmdResNil, errWrongTypeOperation("DECRBY")
	}

	value, _ := obj.Value.(int64)
	if (delta < 0 && value < 0 && delta < (math.MinInt64-value)) ||
		(delta > 0 && value > 0 && delta > (math.MaxInt64-value)) {
		return &CmdRes{R: &wire.Response{
			Value: nil,
		}}, errIntegerOutOfRange
	}

	value += delta
	obj.Value = value
	
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: value},
	}}, nil
}


