// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package commandhandler

import (
	"context"
	"log/slog"

	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/store"
)

// This file is utilized by the CommandHandler to decompose commands that need to be executed
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
func (h *BaseCommandHandler) decomposeRename(ctx context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	// Waiting for GET command response
	var val string
	select {
	case <-ctx.Done():
		slog.Error("CommandHandler timed out waiting for response from shards", slog.String("id", h.id), slog.Any("error", ctx.Err()))
	case preProcessedResp, ok := <-h.preprocessingChan:
		if ok {
			evalResp := preProcessedResp.EvalResponse
			if evalResp.Error != nil {
				return nil, evalResp.Error
			}

			switch v := evalResp.Result.(type) {
			case string:
				val = v
			default:
				return nil, diceerrors.ErrGeneral("no such key")
			}
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
func (h *BaseCommandHandler) decomposeCopy(ctx context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	// Waiting for GET command response
	var resp *ops.StoreResponse
	select {
	case <-ctx.Done():
		slog.Error("CommandHandler timed out waiting for response from shards", slog.String("id", h.id), slog.Any("error", ctx.Err()))
	case preProcessedResp, ok := <-h.preprocessingChan:
		if ok {
			resp = preProcessedResp
		}
	}

	if resp.EvalResponse.Error != nil || resp.EvalResponse.Result == clientio.IntegerZero {
		return nil, &diceerrors.PreProcessError{Result: clientio.IntegerZero}
	}

	if len(cd.Args) < 2 {
		return nil, diceerrors.ErrWrongArgumentCount("COPY")
	}

	newObj, ok := resp.EvalResponse.Result.(*object.InternalObj)
	if !ok {
		return nil, diceerrors.ErrInternalServer
	}

	decomposedCmds := []*cmd.DiceDBCmd{
		{
			Cmd:          "OBJECTCOPY",
			Args:         cd.Args[1:],
			InternalObjs: []*object.InternalObj{newObj},
		},
	}

	return decomposedCmds, nil
}

// decomposePFMerge decomposes the PFMERGE command into individual GET commands for each HLL.
// For each key it creates a separate GET command to get the value at the given key, and waits for all responses to be
// returned before proceeding.
func (h *BaseCommandHandler) decomposePFMerge(ctx context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	// Waiting for GET command response for all the keys to be merged
	resp := make([]*object.InternalObj, 0, len(cd.Args)-1)
	for i := 1; i < len(cd.Args); i++ {
		select {
		case <-ctx.Done():
			slog.Error("CommandHandler timed out waiting for response from shards", slog.String("id", h.id), slog.Any("error", ctx.Err()))
		case preProcessedResp, ok := <-h.preprocessingChan:
			if ok {
				if preProcessedResp.EvalResponse.Error != nil {
					return nil, diceerrors.ErrInvalidHyperLogLogKey
				}

				var hllObjectResult *object.InternalObj
				if preProcessedResp.EvalResponse.Result == clientio.IntegerZero {
					hllObjectResult = nil
				} else {
					hllObjectResult = preProcessedResp.EvalResponse.Result.(*object.InternalObj)
				}

				resp = append(resp, hllObjectResult)
			}
		}
	}

	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args)-1)
	key := cd.Args[0]

	decomposedCmds = append(decomposedCmds,
		&cmd.DiceDBCmd{
			Cmd:          store.PFMERGE,
			Args:         []string{key},
			InternalObjs: resp,
		},
	)

	return decomposedCmds, nil
}

// decomposeMSet decomposes the MSET (Multi-set) command into individual SET commands.
// It expects an even number of arguments (key-value pairs). For each pair, it creates
// a separate SET command to store the value at the given key.
func (h *BaseCommandHandler) decomposeMSet(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
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
func (h *BaseCommandHandler) decomposeMGet(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
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

func (h *BaseCommandHandler) decomposeSInter(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args) < 1 {
		return nil, diceerrors.ErrWrongArgumentCount("SINTER")
	}
	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args))
	for i := 0; i < len(cd.Args); i++ {
		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.Smembers,
				Args: []string{cd.Args[i]},
			},
		)
	}
	return decomposedCmds, nil
}

func (h *BaseCommandHandler) decomposeSDiff(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args) < 1 {
		return nil, diceerrors.ErrWrongArgumentCount("SDIFF")
	}
	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args))
	for i := 0; i < len(cd.Args); i++ {
		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.Smembers,
				Args: []string{cd.Args[i]},
			},
		)
	}
	return decomposedCmds, nil
}

func (h *BaseCommandHandler) decomposeJSONMget(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args) < 2 {
		return nil, diceerrors.ErrWrongArgumentCount("JSON.MGET")
	}

	pattern := cd.Args[len(cd.Args)-1]

	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args))
	for i := 0; i < len(cd.Args)-1; i++ {
		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.JSONGet,
				Args: []string{cd.Args[i], pattern},
			},
		)
	}
	return decomposedCmds, nil
}

func (h *BaseCommandHandler) decomposeTouch(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args) == 0 {
		return nil, diceerrors.ErrWrongArgumentCount("TOUCH")
	}

	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args))
	for i := 0; i < len(cd.Args); i++ {
		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.SingleShardTouch,
				Args: []string{cd.Args[i]},
			},
		)
	}
	return decomposedCmds, nil
}

func (h *BaseCommandHandler) decomposeDBSize(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args) > 0 {
		return nil, diceerrors.ErrWrongArgumentCount("DBSIZE")
	}

	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args))
	for i := uint8(0); i < uint8(h.shardManager.GetShardCount()); i++ {
		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.SingleShardSize,
				Args: []string{},
			},
		)
	}
	return decomposedCmds, nil
}

func (h *BaseCommandHandler) decomposeKeys(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args) != 1 {
		return nil, diceerrors.ErrWrongArgumentCount("KEYS")
	}

	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args))
	for i := uint8(0); i < uint8(h.shardManager.GetShardCount()); i++ {
		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.SingleShardKeys,
				Args: []string{cd.Args[0]},
			},
		)
	}
	return decomposedCmds, nil
}

func (h *BaseCommandHandler) decomposeFlushDB(_ context.Context, cd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error) {
	if len(cd.Args) > 1 {
		return nil, diceerrors.ErrWrongArgumentCount("FLUSHDB")
	}

	decomposedCmds := make([]*cmd.DiceDBCmd, 0, len(cd.Args))
	for i := uint8(0); i < uint8(h.shardManager.GetShardCount()); i++ {
		decomposedCmds = append(decomposedCmds,
			&cmd.DiceDBCmd{
				Cmd:  store.FlushDB,
				Args: cd.Args,
			},
		)
	}
	return decomposedCmds, nil
}
