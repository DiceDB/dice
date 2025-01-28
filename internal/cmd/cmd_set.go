package cmd

import (
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dice/wire"
)

const (
	EX      = "EX"
	PX      = "PX"
	EXAT    = "EXAT"
	PXAT    = "PXAT"
	XX      = "XX"
	NX      = "NX"
	KEEPTTL = "KEEPTTL"
)

var cSET = &DiceDBCommand{
	Name:      "SET",
	HelpShort: "SET puts a new <key, value> pair. If the key already exists then the value will be overwritten.",
	Eval:      evalSET,
}

func init() {
	commandRegistry.AddCommand(cSET)
}

func evalSET(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) <= 1 {
		return cmdResNil, errWrongArgumentCount("SET")
	}

	var key, value string
	var crExistingKey *CmdRes
	var exDurationMs int64 = -1
	var keepttl bool
	var maxExDuration int64 = 9223372036854775

	key, value = c.C.Args[0], c.C.Args[1]
	for i := 2; i < len(c.C.Args); i++ {
		arg := strings.ToUpper(c.C.Args[i])
		switch arg {
		case EX, PX:
			i++
			if i == len(c.C.Args) {
				return cmdResNil, errInvalidSyntax("SET")
			}

			exDuration, err := strconv.ParseInt(c.C.Args[i], 10, 64)
			if err != nil {
				return cmdResNil, errInvalidValue("SET", "EX/PX")
			}

			if exDuration <= 0 || exDuration >= maxExDuration {
				return cmdResNil, errInvalidValue("SET", "EX/PX")
			}

			// converting seconds to milliseconds
			if arg == EX {
				exDuration *= 1000
			}
			exDurationMs = exDuration
		case EXAT, PXAT:
			i++
			if i == len(c.C.Args) {
				return cmdResNil, errInvalidSyntax("SET")
			}

			exDuration, err := strconv.ParseInt(c.C.Args[i], 10, 64)
			if err != nil {
				return cmdResNil, errInvalidValue("SET", "EXAT/PXAT")
			}

			if exDuration < 0 {
				return cmdResNil, errInvalidValue("SET", "EXAT/PXAT")
			}

			if arg == EXAT {
				exDuration *= 1000
			}
			exDurationMs = exDuration - utils.GetCurrentTime().UnixMilli()

			if exDurationMs < 0 {
				exDurationMs = 0
			}
		case XX:
			obj := s.Get(key)
			if obj == nil {
				return cmdResNil, nil
			}
		case NX:
			obj := s.Get(key)
			if obj != nil {
				return cmdResNil, nil
			}
		case KEEPTTL:
			keepttl = true
		case "GET":
			crg, err := evalGET(&Cmd{
				C: &wire.Command{
					Cmd:  "GET",
					Args: []string{key},
				},
			}, s)
			if err != nil {
				return cmdResNil, err
			}
			crExistingKey = crg
		default:
			return cmdResNil, errInvalidSyntax("SET")
		}
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		s.Put(key, s.NewObj(intValue, exDurationMs, object.ObjTypeInt), dstore.WithKeepTTL(keepttl))
	} else {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err == nil {
			s.Put(key, s.NewObj(floatValue, exDurationMs, object.ObjTypeFloat), dstore.WithKeepTTL(keepttl))
		} else {
			s.Put(key, s.NewObj(value, exDurationMs, object.ObjTypeString), dstore.WithKeepTTL(keepttl))
		}
	}

	if crExistingKey != nil {
		return crExistingKey, nil
	}
	return cmdResOK, nil
}
