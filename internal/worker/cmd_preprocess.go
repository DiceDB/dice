package worker

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"
)

func preProcessRename(w *BaseWorker, diceDBCmd *cmd.DiceDBCmd) {
	var rc chan *ops.StoreOp
	var sid shard.ShardID

	key := diceDBCmd.Args[0]
	sid, rc = w.shardManager.GetShardInfo(key)

	preCmd := cmd.DiceDBCmd{
		RequestID:        diceDBCmd.RequestID,
		Cmd:              "GET",
		Args:             []string{key},
		PreProcessingReq: true,
	}

	rc <- &ops.StoreOp{
		SeqID:     0,
		RequestID: preCmd.RequestID,
		Cmd:       &preCmd,
		WorkerID:  w.id,
		ShardID:   sid,
		Client:    nil,
	}
}

func preProcessCopy(w *BaseWorker, diceDBCmd *cmd.DiceDBCmd) {
	var rc chan *ops.StoreOp
	var sid shard.ShardID

	key := diceDBCmd.Args[0]
	sid, rc = w.shardManager.GetShardInfo(key)

	preCmd := cmd.DiceDBCmd{
		RequestID:        diceDBCmd.RequestID,
		Cmd:              "GET",
		Args:             []string{key},
		PreProcessingReq: true,
	}

	rc <- &ops.StoreOp{
		SeqID:     0,
		RequestID: preCmd.RequestID,
		Cmd:       &preCmd,
		WorkerID:  w.id,
		ShardID:   sid,
		Client:    nil,
	}
}
