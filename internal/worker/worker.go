package worker

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/clientio/iohandler"
	"github.com/dicedb/dice/internal/clientio/requestparser"
	"github.com/dicedb/dice/internal/id"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"
)

// Worker interface
type Worker interface {
	ID() string
	Start(ctx context.Context) error
	Stop() error
}

type BaseWorker struct {
	id           string
	ioHandler    iohandler.IOHandler
	parser       requestparser.Parser
	shardManager *shard.ShardManager
	respChan     chan *ops.StoreResponse
}

func NewWorker(wid string, respChan chan *ops.StoreResponse, ioHandler iohandler.IOHandler, parser requestparser.Parser, shardManager *shard.ShardManager) *BaseWorker {
	return &BaseWorker{
		id:           wid,
		ioHandler:    ioHandler,
		parser:       parser,
		shardManager: shardManager,
		respChan:     respChan,
	}
}

func (w *BaseWorker) ID() string {
	return w.id
}

func (w *BaseWorker) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			data, err := w.ioHandler.ReadRequest(ctx)
			if err != nil {
				log.Warnf("Error handling client requests for worker: %s, %v", w.id, err)
				continue
			}

			cmds, err := w.parser.Parse(data)
			if err != nil {
				err = w.ioHandler.WriteResponse(ctx, clientio.Encode(err, true))
				if err != nil {
					log.Warnf("Error writing response for worker: %s, %v", w.id, err)
				}
			}

			if len(cmds) == 0 {
				err = w.ioHandler.WriteResponse(ctx, clientio.Encode("ERR: Invalid request", true))
			}

			// DiceDB supports clients to send only one request at a time
			// We also need to ensure that the client is blocked until the response is received
			if len(cmds) > 1 {
				err = w.ioHandler.WriteResponse(ctx, clientio.Encode("ERR: Multiple commands not supported", true))
				if err != nil {
					log.Warnf("Error writing response for worker: %s, %v", w.id, err)
				}
			}

			request := cmds[0]
			reqID := id.NextID()
			sReq := &ops.StoreOp{
				RequestID: reqID,
				Cmd:       request,
				WorkerID:  w.id,
			}

			err = w.SendToShards(ctx, sReq)
			if err != nil {
				log.Warnf("Error sending request to shards: %v", err)
				err = w.ioHandler.WriteResponse(ctx, clientio.Encode("ERR: Internal server error, unable to process request", true))
				if err != nil {
					log.Warnf("Error writing response for worker: %s, %v", w.id, err)
				}
			}

			shardResp := <-w.respChan

			err = w.ioHandler.WriteResponse(ctx, shardResp.Result)
			if err != nil {
				log.Warnf("Error writing response for worker: %s, %v", w.id, err)
			}
		}
	}
}

func (w *BaseWorker) SendToShards(ctx context.Context, request *ops.StoreOp) error {
	w.shardManager.GetShard(0).ReqChan <- request

	// Implement request handling logic
	// This involves sending the request to the appropriate shard
	// This involves extensive work where we need to find the key in the request.
	// We need to invoke EvalXXX() functions to find the key/keys in the command and send it to the appropriate shard
	return nil
}

func (w *BaseWorker) Stop() error {
	// Implement worker shutdown logic
	return nil
}
