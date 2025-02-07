package cmd

import (
	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

var cGETDEL = &DiceDBCommand{
	Name:      "GETDEL",
	HelpShort: "GETDEL returns the value for the queried key in args.",
	Eval:      evalGETDEL,
}

func init() {
	commandRegistry.AddCommand(cGETDEL)
}

// evalGETDEL returns the value for the queried key in args
//
// The function expects exactly one argument: the key to get.
// If the key exists, it will be deleted before its value is returned.
// evalGETDEL returns cmdResNil if key is expired or it does not exist
//
// Parameters:
//   - c *Cmd: The command context containing the arguments
//   - s *dstore.Store: The data store instance
//
// Returns:
//   - *CmdRes: Response containing the value for the queried key.
//   - error: Error if wrong number of arguments or wrong value type.
func evalGETDEL(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("GETDEL")
	}

	key := c.C.Args[0]

	// Getting the key based on previous touch value
	obj := s.GetNoTouch(key)

	// If key does not exist, return nil
	if obj == nil {
		return cmdResNil, nil
	}

	// If the object exists, check if it is either a JSON or SET object.
	switch obj.Type {
	case object.ObjTypeSet, object.ObjTypeJSON:
		return cmdResNil, errWrongTypeOperation("GETDEL")
	}

	// If the object exists, check if it is a Set object.
	// Get the key from the hash table
	objVal := s.GetDel(key)

	// Decode and return the value based on its encoding
	switch oType := objVal.Type; oType {
	case object.ObjTypeInt:
		// Value is stored as an int64, so use type assertion
		if IsInt64(objVal.Value) {
			return &CmdRes{R: &wire.Response{
				Value: &wire.Response_VInt{VInt: objVal.Value.(int64)},
			}}, nil
		} else if IsString(objVal.Value) {
			return cmdResNil, errUnexpectedType("int64", "string")
		} else {
			return cmdResNil, errUnexpectedType("int64", "unknown")
		}
	case object.ObjTypeString:
		// Value is stored as a string, use type assertion
		if IsString(objVal.Value) {
			return &CmdRes{R: &wire.Response{
				Value: &wire.Response_VStr{VStr: objVal.Value.(string)},
			}}, nil
		} else if IsInt64(objVal.Value) {
			return cmdResNil, errUnexpectedType("string", "int64")
		} else {
			return cmdResNil, errUnexpectedType("string", "unknown")
		}
	case object.ObjTypeByteArray:
		// Value is stored as a bytearray, use type assertion
		if val, ok := objVal.Value.([]byte); ok {
			return &CmdRes{R: &wire.Response{
				Value: &wire.Response_VStr{VStr: string(val)},
			}}, nil
		}
		return cmdResNil, errWrongTypeOperation("GETDEL")
	default:
		return cmdResNil, errWrongTypeOperation("GETDEL")
	}
}
