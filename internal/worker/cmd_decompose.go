package worker

import (
	"context"
	"log/slog"

	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/store"
)

// This file is utilized by the Worker to decompose commands that need to be executed
// across multiple shards. For commands that operate on multiple keys or necessitate
// distribution across shards (e.g., MultiShard commands), a Breakup function is invoked
// to transform the original command into multiple smaller commands, each directed at
// a specific shard.
//
// Each Breakup function processes the input command, identifies relevant keys and their
// corresponding shards, and generates a set of individual commands. These commands are
// sent to the appropriate shards for parallel execution.

// decomposeRename breaks down the RENAME command into separate DELETE and SET commands.
// It first waits for the result of a GET command from shards. If successful, it removes
// the old key using a DEL command and sets the new key with the retrieved value using a SET command.
func decomposeRename(ctx context.Context, w *BaseWorker, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	// Waiting for GET command response
	var val string
	select {
	case <-ctx.Done():
		slog.Error("Timed out waiting for response from shards", slog.String("workerID", w.id), slog.Any("error", ctx.Err()))
	case preProcessedResp, ok := <-w.preprocessingChan:
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
			Cmd:  store.Del,
			Args: []string{cd.Args[0]},
		},
		&cmd.DiceDBCmd{
			Cmd:  store.Set,
			Args: []string{cd.Args[1], val},
		},
	)

	return decomposedCmds, nil
}

// decomposeCopy breaks down the COPY command into a SET command that copies a value from
// one key to another. It first retrieves the value of the original key from shards, then
// sets the value to the destination key using a SET command.
func decomposeCopy(ctx context.Context, w *BaseWorker, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	// Waiting for GET command response
	var val string
	select {
	case <-ctx.Done():
		slog.Error("Timed out waiting for response from shards", slog.String("workerID", w.id), slog.Any("error", ctx.Err()))
	case preProcessedResp, ok := <-w.preprocessingChan:
		if ok {
			evalResp := preProcessedResp.EvalResponse
			if evalResp.Error != nil {
				return nil, evalResp.Error
			}

			val = evalResp.Result.(string)
		}
	}

	if len(cd.Args) != 2 {
		return nil, diceerrors.ErrWrongArgumentCount("COPY")
	}

	decomposedCmds := []*cmd.DiceDBCmd{}
	decomposedCmds = append(decomposedCmds,
		&cmd.DiceDBCmd{
			Cmd:  store.Set,
			Args: []string{cd.Args[1], val},
		},
	)

	return decomposedCmds, nil
}

// decomposeMSet decomposes the MSET (Multi-set) command into individual SET commands.
// It expects an even number of arguments (key-value pairs). For each pair, it creates
// a separate SET command to store the value at the given key.
func decomposeMSet(_ context.Context, _ *BaseWorker, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args)%2 != 0 {
		return nil, diceerrors.ErrWrongArgumentCount("MSET")
	}

	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args)/2)

	for i := 0; i < len(cd.Args)-1; i += 2 {
		key := cd.Args[i]
		val := cd.Args[i+1]

		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.Set,
				Args: []string{key, val},
			},
		)
	}
	return decomposedCmds, nil
}

// decomposeMGet decomposes the MGET (Multi-get) command into individual GET commands.
// It expects a list of keys, and for each key, it creates a separate GET command to
// retrieve the value associated with that key.
func decomposeMGet(_ context.Context, _ *BaseWorker, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args) < 1 {
		return nil, diceerrors.ErrWrongArgumentCount("MGET")
	}
	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args))
	for i := 0; i < len(cd.Args); i++ {
		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.Get,
				Args: []string{cd.Args[i]},
			},
		)
	}
	return decomposedCmds, nil
}
