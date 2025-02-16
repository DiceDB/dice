package cmd

import (
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/dicedb/dicedb-go/wire"
)

var cExpireTime *DiceDBCommand = &DiceDBCommand{
	Name:      "EXPIRETIME",
	HelpShort: `EXPIRETIME returns the absolute Unix timestamp in seconds at which the given key will expire`,
	Eval:      evalExpireTime,
}

func init() {
	commandRegistry.AddCommand(cExpireTime)
}

// evalEXPIRETIME returns the absolute Unix timestamp (since January 1, 1970) in seconds at which the given key will expire
// args should contain only 1 value, the key
// Returns expiration Unix timestamp in seconds.
// Returns -1 if the key exists but has no associated expiration time.
// Returns -2 if the key does not exist.
func evalExpireTime(c *Cmd, dst *dstore.Store) (*CmdRes, error) {
	// check for correct number of arguments
	if len(c.C.Args) != 1 {
		return cmdResNil, errWrongArgumentCount("EXPIRETIME")
	}

	key := c.C.Args[0]
	obj := dst.Get(key)

	// returns -2 as the object is unavailable
	if obj == nil {
		return cmdResIntNegTwo, nil
	}

	getExpiry, ok := dstore.GetExpiry(obj, dst)

	// returns -1 as the key doesn't have an expiration time set
	if !ok {
		return cmdResIntNegOne, nil
	}

	// returns the absolute Unix timestamp (since January 1, 1970) in seconds at which the given key will expire
	return &CmdRes{R: &wire.Response{
		Value: &wire.Response_VInt{
			VInt: int64(getExpiry / 1000),
		},
	}}, nil
}
