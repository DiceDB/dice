package cmd

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/dicedb/dice/internal/object"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

var cGET = &DiceDBCommand{
	Name:      "GET",
	HelpShort: "GET returns the value for the key in args",
	Eval:      evalGET,
}

func init() {
	commandRegistry.AddCommand(cGET)
}

func evalGET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("GET")
	}
	key := c.C.Args[0]
	obj := s.Get(key)
	fmt.Println(obj)
	if obj == nil {
		return cmdResNil, nil
	}

	switch obj.Type {
	case object.ObjTypeInt:
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VInt{VInt: obj.Value.(int64)},
		}, Mu: &sync.Mutex{}}, nil
	case object.ObjTypeString:
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VStr{VStr: obj.Value.(string)},
		}, Mu: &sync.Mutex{}}, nil
	case object.ObjTypeByteArray, object.ObjTypeHLL:
		return &CmdRes{R: &wire.Response{
			Value: &wire.Response_VBytes{VBytes: obj.Value.([]byte)},
		}, Mu: &sync.Mutex{}}, nil
	default:
		slog.Error("unknown object type", "type", obj.Type)
		return cmdResNil, errUnknownObjectType
	}
}
