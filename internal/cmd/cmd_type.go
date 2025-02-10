package cmd

import (
	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cType = &DiceDBCommand{
	Name:	   "TYPE",
	HelpShort: "Determine data type of the value stored at a specified key",
	Eval:		evalTYPE,
}

func init() {
	commandRegistry.AddCommand(cType)
}

func evalTYPE(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("TYPE")
	}

	key := c.C.Args[0]
	obj := s.Get(key)

	if obj == nil {
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VStr{
				VStr: "none",
			},
		}}, nil
	}

	var typeStr string
	switch oType := obj.Type; oType {
	case object.ObjTypeString, object.ObjTypeInt, object.ObjTypeByteArray:
		typeStr = "string"
	case object.ObjTypeDequeue:
		typeStr = "list"
	case object.ObjTypeSet:
		typeStr = "set"
	case object.ObjTypeHashMap:
		typeStr = "hash"
	case object.ObjTypeSortedSet:
		typeStr = "zset"
	default:
		typeStr = "non-supported type"
	}

	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VStr{
			VStr: typeStr,
		},
	}}, nil
}