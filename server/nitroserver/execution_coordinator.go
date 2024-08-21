package nitroserver

import (
	"context"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/core"
)

var diceCores int
var spool *ShardPool

func InitShards(ctx context.Context, wg *sync.WaitGroup, cores int) {
	spool = NewShardPool(ctx, wg, cores)
	diceCores = cores
}

func GetDiceCores() int {
	return diceCores
}

func SubmitAndListenClientOperation(comm *core.Client, cmds core.RedisCmds) {
	iothread := NewIOThread()
	go iothread.Run(spool)

	go func() {
		var executionResponses [][]byte
		var executionCommand *core.RedisCmd
		var respCount int

		for {
			ioresult, ok := <-iothread.ioresch

			respCount++
			if !ok {
				log.Info("ioresch channel closed. Exiting goroutine.")
				return
			}

			executionResponses = append(executionResponses, ioresult.message)
			executionCommand = ioresult.cmd

			// Manual break when we have read from all expected shards
			if respCount == ioresult.targetExecutionShards {
				break
			}

			// Ideally this shouldn't happen. Just a guard rail where we have read all shards
			// and ioresult.targetExecutionShards is not set for some reason.
			if respCount == diceCores {
				break
			}
		}

		clientResponse := ReduceShardedResponse(executionResponses, executionCommand)

		if _, err := comm.Write(clientResponse); err != nil {
			log.Info("Error writing to client")
			log.Error(err)
		}
	}()

	for _, cmd := range cmds {
		iothread.ioreqch <- &IORequest{conn: comm, cmd: cmd, keys: extractKeysForOperation(*cmd)}
	}
}

// Identifies the Key used in the operation.
// Otherwise return emptyslice if operation is not bound to a key.
// This func will need refinement as we support more commands.
func extractKeysForOperation(cmd core.RedisCmd) []string {
	var keys = []string{}

	if cmd.Cmd == "DEL" {
		keys = append(keys, cmd.Args...)
	} else {
		keyBasedCommands := []string{
			"SET", "GET", "TTL", "EXPIRE", "INCR", "QINTINS", "QINTLEN", "QINTPEEK",
			"QINTREM", "BFINIT", "BFADD", "BFEXISTS", "BFINFO", "QREFINS", "QREFREM",
			"QREFLEN", "QREFLEN", "QREFPEEK", "STACKINTPUSH", "STACKINTPOP", "STACKINTLEN",
			"STACKINTPEEK", "STACKREFPUSH", "STACKREFPOP", "STACKREFLEN", "STACKREFPEEK",
		}

		for _, b := range keyBasedCommands {
			if b == cmd.Cmd {
				keys = append(keys, cmd.Args[0])
			}
		}
	}

	return keys
}

func ReduceShardedResponse(responses [][]byte, cmd *core.RedisCmd) []byte {
	// Just returning any Non-Nil value from the responses atm.
	// Ideally this would be a reduce operation that is applicable
	// for the concerned RedisCmd command used to achieve the result.

	var reducedResult []byte = responses[0]
	var emptyResponse = string([]byte{36, 45, 49, 13, 10})

	for _, resp := range responses {
		if string(resp) != emptyResponse {
			reducedResult = resp
		}
	}
	return reducedResult
}
