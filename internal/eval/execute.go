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

type Eval struct {
	cmd                   *cmd.DiceDBCmd
	client                *comm.Client
	store                 *dstore.Store
	isHTTPOperation       bool
	isWebSocketOperation  bool
	isPreprocessOperation bool
}

func NewEval(c *cmd.DiceDBCmd, client *comm.Client, store *dstore.Store, httpOp bool, websocketOp bool, preProcessing bool) *Eval {
	return &Eval{
		cmd:                   c,
		client:                client,
		store:                 store,
		isHTTPOperation:       httpOp,
		isWebSocketOperation:  websocketOp,
		isPreprocessOperation: preProcessing,
	}
}

func (e *Eval) PreProcessCommand() *EvalResponse {
	if f, ok := PreProcessing[e.cmd.Cmd]; ok {
		return f(e.cmd, e.store)
	}
	return &EvalResponse{Result: nil, Error: diceerrors.ErrInternalServer}
}

func (e *Eval) ExecuteCommand() *EvalResponse {
	diceCmd, ok := DiceCmds[e.cmd.Cmd]
	if !ok {
		return &EvalResponse{Result: diceerrors.NewErrWithFormattedMessage("unknown command '%s', with args beginning with: %s", e.cmd.Cmd, strings.Join(e.cmd.Args, " ")), Error: nil}
	}

	// Temporary logic till we move all commands to new eval logic.
	// MigratedDiceCmds map contains refactored eval commands
	// For any command we will first check in the existing map
	// if command is NA then we will check in the new map

	if diceCmd.IsMigrated {
		return diceCmd.NewEval(e.cmd, e.store)
	}

	// The following commands could be handled at the shard level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	switch diceCmd.Name {
	// Old implementation kept as it is, but we will be moving
	// to the new implementation soon for all commands
	case "SUBSCRIBE", "Q.WATCH":
		return &EvalResponse{Result: EvalQWATCH(e.cmd.Args, e.isHTTPOperation, e.isWebSocketOperation, e.client, e.store), Error: nil}
	case "UNSUBSCRIBE", "Q.UNWATCH":
		return &EvalResponse{Result: EvalQUNWATCH(e.cmd.Args, e.isHTTPOperation, e.client), Error: nil}
	case auth.Cmd:
		return &EvalResponse{Result: EvalAUTH(e.cmd.Args, e.client), Error: nil}
	case "ABORT":
		return &EvalResponse{Result: clientio.RespOK, Error: nil}
	default:
		return &EvalResponse{Result: diceCmd.Eval(e.cmd.Args, e.store), Error: nil}
	}
}
