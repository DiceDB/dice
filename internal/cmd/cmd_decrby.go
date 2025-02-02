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
	
	decrAmount, err := strconv.ParseInt(c.C.Args[1], 10, 64)
	if err != nil {
		return nil, errIntegerOutOfRange
	}
	return incrDecrCmd(c.C.Args, -decrAmount, s)
}

func incrDecrCmd(args []string, incr int64, store *dstore.Store) (*CmdRes, error) {
	key := args[0]
	obj := store.Get(key)
	if obj == nil {
		obj = store.NewObj(incr, -1, object.ObjTypeInt)
		store.Put(key, obj)
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: incr},
		}}, nil
	}

	switch obj.Type {
	case object.ObjTypeString:
		return cmdResNil, errIntegerOutOfRange
	case object.ObjTypeInt:
		return cmdResNil, errWrongTypeOperation("DECRBY")
	}
	
	i, _ := obj.Value.(int64)
	if (incr < 0 && i < 0 && incr < (math.MinInt64-i)) ||
		(incr > 0 && i > 0 && incr > (math.MaxInt64-i)) {
		return &CmdRes{R: &wire.Response{
			Value: nil,
		}}, errIntegerOutOfRange
	}

	i += incr
	obj.Value = i
	return	&CmdRes{R: &wire.Response{	
		Value: &wire.Response_VInt{VInt: i},
	}}, nil
}