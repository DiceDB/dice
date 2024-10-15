package worker

import (
	"context"
	"log/slog"

	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
)

// Breakup file is used by Worker to split commands that need to be executed
// across multiple shards. For commands that operate on multiple keys or
// require distribution across shards (e.g., MultiShard commands), a Breakup
// function is invoked to break the original command into multiple smaller
// commands, each targeted at a specific shard.
//
// Each Breakup function takes the input command, identifies the relevant keys
// and their corresponding shards, and generates a set of commands that are
// individually sent to the respective shards. This ensures that commands can
// be executed in parallel across shards, allowing for horizontal scaling
// and distribution of data processing.
//
// The result is a list of commands, one for each shard, which are then
// scattered to the shard threads for execution.
func decomposeRename(ctx context.Context, w *BaseWorker, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	// Waiting for GET command response
	var val string
	select {
	case <-ctx.Done():
		w.logger.Error("Timed out waiting for response from shards", slog.String("workerID", w.id), slog.Any("error", ctx.Err()))
	case preProcessedResp, ok := <-w.preProcessingChan:
		if ok {
			evalResp := preProcessedResp.EvalResponse
			if evalResp.Error != nil {
				return nil, evalResp.Error
			}

			val = evalResp.Result.(string)
		}
	}

	if len(cd.Args) != 2 {
		return nil, diceerrors.ErrWrongArgumentCount("RENAME")
	}

	decomposedCmds := []*cmd.DiceDBCmd{}
	decomposedCmds = append(decomposedCmds,
		&cmd.DiceDBCmd{
			RequestID: cd.RequestID,
			Cmd:       "DEL",
			Args:      []string{cd.Args[0]},
		},
		&cmd.DiceDBCmd{
			RequestID: cd.RequestID,
			Cmd:       "SET",
			Args:      []string{cd.Args[1], val},
		},
	)

	return decomposedCmds, nil
}

func decomposeCopy(ctx context.Context, w *BaseWorker, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	// Waiting for GET command response
	var val string
	select {
	case <-ctx.Done():
		w.logger.Error("Timed out waiting for response from shards", slog.String("workerID", w.id), slog.Any("error", ctx.Err()))
	case preProcessedResp, ok := <-w.preProcessingChan:
		if ok {
			evalResp := preProcessedResp.EvalResponse
			if evalResp.Error != nil {
				return nil, evalResp.Error
			}

			val = evalResp.Result.(string)
		}
	}

	if len(cd.Args) != 2 {
		return nil, diceerrors.ErrWrongArgumentCount("RENAME")
	}

	decomposedCmds := []*cmd.DiceDBCmd{}
	decomposedCmds = append(decomposedCmds,
		&cmd.DiceDBCmd{
			RequestID: cd.RequestID,
			Cmd:       "SET",
			Args:      []string{cd.Args[1], val},
		},
	)

	return decomposedCmds, nil
}
