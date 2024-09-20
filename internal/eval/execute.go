package eval

import (
	"fmt"
	"strings"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	dstore "github.com/dicedb/dice/internal/store"
)

func ExecuteCommand(cmd *cmd.RedisCmd, client *comm.Client, store *dstore.Store, http bool) EvalResponse {
	diceCmd, ok := DiceCmds[cmd.Cmd]
	if !ok {
		return EvalResponse{Result: nil, Error: fmt.Errorf("unknown command '%s', with args beginning with: %s", cmd.Cmd, strings.Join(cmd.Args, " "))}
	}

	// Till the time we refactor to handle QWATCH differently using HTTP Streaming/SSE
	if http {
		return EvalResponse{Result: diceCmd.Eval(cmd.Args, store), Error: nil}
	}

	// Temporary logic till we move all commands to new eval logic.
	// MigratedDiceCmds map contains refactored eval commands
	// For any command we will first check in the exisiting map
	// if command is NA then we will check in the new map
	if diceCmd.IsMigrated {
		return diceCmd.NewEval(cmd.Args, store)
	}

	// The following commands could be handled at the shard level, however, we can randomly let any shard handle them
	// to reduce load on main server.
	switch diceCmd.Name {
	// Old implementation kept as it is, but we will be moving
	// to the new implmentation soon for all commands
	case "SUBSCRIBE", "QWATCH":
		return EvalResponse{Result: EvalQWATCH(cmd.Args, client.Fd, store), Error: nil}
	case "UNSUBSCRIBE", "QUNWATCH":
		return EvalResponse{Result: EvalQUNWATCH(cmd.Args, client.Fd), Error: nil}
	case auth.AuthCmd:
		return EvalResponse{Result: EvalAUTH(cmd.Args, client), Error: nil}
	case "ABORT":
		return EvalResponse{Result: clientio.RespOK, Error: nil}
	default:
		return EvalResponse{Result: diceCmd.Eval(cmd.Args, store), Error: nil}
	}
}
