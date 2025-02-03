package cmd

import (
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

// evalDECR decrements an integer value stored at the specified key by 1.
//
// The function expects exactly one argument: the key to decrement.
// If the key does not exist, it is initialized with value -1.
// If the key exists but does not contain an integer, an error is returned.
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: Response containing the new integer value after decrement
//   - error: Error if wrong number of arguments or wrong value type
func evalDECR(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("DECR")
	}

	delta := int64(-1)

	key := c.C.Args[0]
	obj := s.Get(key)
	if obj == nil {
		obj = s.NewObj(delta, INFINITE_EXPIRATION, object.ObjTypeInt)
		s.Put(key, obj)
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: delta},
		}}, nil
	}

	switch obj.Type {
	case object.ObjTypeInt:
		break
	default:
		return cmdResNil, errWrongTypeOperation("DECR")
	}

	val, _ := obj.Value.(int64)
	val += delta

	obj.Value = val
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{VInt: val},
	}}, nil
}
