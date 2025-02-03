package cmd

import (
	"strconv"

	dstore "github.com/dicedb/dice/internal/store"
	"google.golang.org/protobuf/types/known/structpb"
)

var cGETWATCH = &DiceDBCommand{
	Name:      "GET.WATCH",
	HelpShort: "GET.WATCH creates a query subscription over the GET command",
	Eval:      evalGETWATCH,
}

func init() {
	commandRegistry.AddCommand(cGETWATCH)
}

func evalGETWATCH(c *Cmd, s *dstore.Store) (*CmdRes, error) {
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("GET.WATCH")
	}

	r, err := evalGET(c, s)
	if err != nil {
		return nil, err
	}

	if r.R.Attrs == nil {
		r.R.Attrs = &structpb.Struct{
			Fields: make(map[string]*structpb.Value),
		}
	}

	r.R.Attrs.Fields["fingerprint"] = structpb.NewStringValue(strconv.FormatUint(uint64(c.GetFingerprint()), 10))
	return r, nil
}
