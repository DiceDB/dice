package worker

import (
	"context"
	"log/slog"

	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"
)

// Gather file is used by Worker to collect and process responses
// from multiple shards. For commands that are executed across
// several shards (e.g., MultiShard commands), a Gather function
// is responsible for aggregating the results.
//
// Each Gather function takes input in the form of shard responses,
// applies command-specific logic to combine or process these
// individual shard responses, and returns the final response
// expected by the client.
//
// The result is a unified response that reflects the combined
// outcome of operations executed across multiple shards, ensuring
// that the client receives a single, cohesive result.
func composeRename(ctx context.Context, originalCmd *cmd.DiceDBCmd, worker *BaseWorker, responses ...eval.EvalResponse) error {
	for idx := range responses {
		if responses[idx].Error != nil {
			err := worker.ioHandler.Write(ctx, diceerrors.ErrInternalServer)
			if err != nil {
				return err
			}
		}
	}

	cmds := cmd.DiceDBCmd{
		RequestID: ctx.Value("request_id").(uint32),
		Cmd:       "SET",
		Args:      []string{originalCmd.Args[1], responses[0].Result.(string)},
	}

	var rc chan *ops.StoreOp
	var sid shard.ShardID
	var key string
	if len(cmds.Args) > 0 {
		key = cmds.Args[0]
	} else {
		key = cmds.Cmd
	}

	sid, rc = worker.shardManager.GetShardInfo(key)

	rc <- &ops.StoreOp{
		SeqID:     0,
		RequestID: cmds.RequestID,
		Cmd:       &cmds,
		WorkerID:  worker.id,
		ShardID:   sid,
		Client:    nil,
	}

	var evalResp []eval.EvalResponse

	select {
	case <-ctx.Done():
		worker.logger.Error("Timed out waiting for response from shards", slog.String("workerID", worker.id), slog.Any("error", ctx.Err()))
	case resp, ok := <-worker.respChan:
		if ok {
			evalResp = append(evalResp, *resp.EvalResponse)
		}
	case sError, ok := <-worker.shardManager.ShardErrorChan:
		if ok {
			worker.logger.Error("Error from shard", slog.String("workerID", worker.id), slog.Any("error", sError))
		}
	}

	if evalResp[0].Error != nil {
		err := worker.ioHandler.Write(ctx, evalResp[0].Error)
		if err != nil {
			worker.logger.Debug("Error sending response to client", slog.String("workerID", worker.id), slog.Any("error", err))
		}
		return err
	}

	err := worker.ioHandler.Write(ctx, evalResp[0].Result)
	if err != nil {
		worker.logger.Debug("Error sending response to client", slog.String("workerID", worker.id), slog.Any("error", err))
		return err
	}

	return nil
}
