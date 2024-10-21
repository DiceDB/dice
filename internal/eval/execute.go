package eval

import (
	"strings"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
)

func ExecuteCommand(c *cmd.DiceDBCmd, client *comm.Client, store *dstore.Store, httpOp, websocketOp bool) *EvalResponse {
	diceCmd, ok := DiceCmds[c.Cmd]
	if !ok {
		return &EvalResponse{Result: diceerrors.NewErrWithFormattedMessage("unknown command '%s', with args beginning with: %s", c.Cmd, strings.Join(c.Args, " ")), Error: nil}
	}

	// Temporary logic till we move all commands to new eval logic.
	// MigratedDiceCmds map contains refactored eval commands
	// For any command we will first check in the existing map
	// if command is NA then we will check in the new map
	if diceCmd.IsMigrated {
		return diceCmd.NewEval(c.Args, store)
	}

	// The following commands could be handled at the shard level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	switch diceCmd.Name {
	// Old implementation kept as it is, but we will be moving
	// to the new implementation soon for all commands
	case "SUBSCRIBE", "Q.WATCH":
		return &EvalResponse{Result: EvalQWATCH(c.Args, httpOp, websocketOp, client, store), Error: nil}
	case "UNSUBSCRIBE", "Q.UNWATCH":
		return &EvalResponse{Result: EvalQUNWATCH(c.Args, httpOp, client), Error: nil}
	case auth.Cmd:
		return &EvalResponse{Result: EvalAUTH(c.Args, client), Error: nil}
	case "ABORT":
		return &EvalResponse{Result: clientio.RespOK, Error: nil}
	default:
		return &EvalResponse{Result: diceCmd.Eval(c.Args, store), Error: nil}
	}
}
