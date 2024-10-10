package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/querymanager"
	"github.com/dicedb/dice/internal/watchmanager"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/clientio/iohandler"
	"github.com/dicedb/dice/internal/clientio/requestparser"
	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"
)

// Worker interface
type Worker interface {
	ID() string
	Start(context.Context) error
	Stop() error
}

type BaseWorker struct {
	id              string
	ioHandler       iohandler.IOHandler
	parser          requestparser.Parser
	shardManager    *shard.ShardManager
	respChan        chan *ops.StoreResponse
	adhocReqChan    chan *cmd.DiceDBCmd
	Session         *auth.Session
	globalErrorChan chan error
	logger          *slog.Logger
}

func NewWorker(wid string, respChan chan *ops.StoreResponse,
	ioHandler iohandler.IOHandler, parser requestparser.Parser,
	shardManager *shard.ShardManager, gec chan error,
	logger *slog.Logger) *BaseWorker {
	return &BaseWorker{
		id:              wid,
		ioHandler:       ioHandler,
		parser:          parser,
		shardManager:    shardManager,
		globalErrorChan: gec,
		respChan:        respChan,
		logger:          logger,
		Session:         auth.NewSession(),
		adhocReqChan:    make(chan *cmd.DiceDBCmd, config.DiceConfig.Server.AdhocReqChanBufSize),
	}
}

func (w *BaseWorker) ID() string {
	return w.id
}

func (w *BaseWorker) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	dataChan := make(chan []byte)
	readErrChan := make(chan error)

	go func() {
		for {
			data, err := w.ioHandler.Read(ctx)
			if err != nil {
				readErrChan <- err
				return
			}
			dataChan <- data
		}
	}()

	for {
		select {
		case <-ctx.Done():
			err := w.Stop()
			if err != nil {
				w.logger.Warn("Error stopping worker:", slog.String("workerID", w.id), slog.Any("error", err))
			}
			return ctx.Err()
		case err := <-errChan:
			if err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
					w.logger.Error("Connection closed for worker", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
			}
			return fmt.Errorf("error writing response: %w", err)
		case cmdReq := <-w.adhocReqChan:
			// Handle adhoc requests of DiceDBCmd
			func() {
				execCtx, cancel := context.WithTimeout(ctx, 6*time.Second) // Timeout set to 6 seconds for integration tests
				defer cancel()

				// adhoc requests should be classified as watch requests
				w.executeCommandHandler(execCtx, errChan, []*cmd.DiceDBCmd{cmdReq}, true)
			}()
		case data := <-dataChan:
			cmds, err := w.parser.Parse(data)
			if err != nil {
				err = w.ioHandler.Write(ctx, err)
				if err != nil {
					w.logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
			}
			if len(cmds) == 0 {
				err = w.ioHandler.Write(ctx, fmt.Errorf("ERR: Invalid request"))
				if err != nil {
					w.logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
				continue
			}

			// DiceDB supports clients to send only one request at a time
			// We also need to ensure that the client is blocked until the response is received
			if len(cmds) > 1 {
				err = w.ioHandler.Write(ctx, fmt.Errorf("ERR: Multiple commands not supported"))
				if err != nil {
					w.logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
			}

			err = w.isAuthenticated(cmds[0])
			if err != nil {
				werr := w.ioHandler.Write(ctx, err)
				if werr != nil {
					w.logger.Debug("Write error, connection closed possibly", slog.Any("error", errors.Join(err, werr)))
					return errors.Join(err, werr)
				}
			}
			// executeCommand executes the command and return the response back to the client
			func(errChan chan error) {
				execCtx, cancel := context.WithTimeout(ctx, 6*time.Second) // Timeout set to 6 seconds for integration tests
				defer cancel()
				w.executeCommandHandler(execCtx, errChan, cmds, false)
			}(errChan)
		case err := <-readErrChan:
			w.logger.Debug("Read error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
			return err
		}
	}
}

func (w *BaseWorker) executeCommandHandler(execCtx context.Context, errChan chan error, cmds []*cmd.DiceDBCmd, isWatchNotification bool) {
	err := w.executeCommand(execCtx, cmds[0], isWatchNotification)
	if err != nil {
		w.logger.Error("Error executing command", slog.String("workerID", w.id), slog.Any("error", err))
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ETIMEDOUT) {
			w.logger.Debug("Connection closed for worker", slog.String("workerID", w.id), slog.Any("error", err))
			errChan <- err
		}
	}
}

func (w *BaseWorker) executeCommand(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, isWatchNotification bool) error {
	// Break down the single command into multiple commands if multisharding is supported.
	// The length of cmdList helps determine how many shards to wait for responses.
	cmdList := make([]*cmd.DiceDBCmd, 0)

	// Retrieve metadata for the command to determine if multisharding is supported.
	meta, ok := CommandsMeta[diceDBCmd.Cmd]
	if !ok {
		// If no metadata exists, treat it as a single command and not migrated
		cmdList = append(cmdList, diceDBCmd)
	} else {
		// Depending on the command type, decide how to handle it.
		switch meta.CmdType {
		case Global:
			// If it's a global command, process it immediately without involving any shards.
			err := w.ioHandler.Write(ctx, meta.WorkerCommandHandler(diceDBCmd.Args))
			w.logger.Debug("Error executing for worker", slog.String("workerID", w.id), slog.Any("error", err))
			return err

		case SingleShard:
			// For single-shard or custom commands, process them without breaking up.
			cmdList = append(cmdList, diceDBCmd)

		case MultiShard:
			// If the command supports multisharding, break it down into multiple commands.
			cmdList = meta.decomposeCommand(diceDBCmd)

		case Custom:
			// if command is of type Custom, write a custom logic around it
			switch diceDBCmd.Cmd {
			case CmdAuth:
				err := w.ioHandler.Write(ctx, w.RespAuth(diceDBCmd.Args))
				if err != nil {
					w.logger.Error("Error sending auth response to worker", slog.String("workerID", w.id), slog.Any("error", err))
				}
				return err
			case CmdAbort:
				err := w.ioHandler.Write(ctx, clientio.OK)
				if err != nil {
					w.logger.Error("Error sending abort response to worker", slog.String("workerID", w.id), slog.Any("error", err))
				}
				w.logger.Info("Received ABORT command, initiating server shutdown", slog.String("workerID", w.id))
				w.globalErrorChan <- diceerrors.ErrAborted
				return err
			default:
				cmdList = append(cmdList, diceDBCmd)
			}

		case Watch:
			// Generate the Cmd being watched. All we need to do is remove the .WATCH suffix from the command and pass
			// it along as is.
			watchCmd := &cmd.DiceDBCmd{
				Cmd:  diceDBCmd.Cmd[:len(diceDBCmd.Cmd)-6], // Remove the .WATCH suffix
				Args: diceDBCmd.Args,
			}
			cmdList = append(cmdList, watchCmd)
			isWatchNotification = true
		}
	}

	// Scatter the broken-down commands to the appropriate shards.
	if err := w.scatter(ctx, cmdList); err != nil {
		return err
	}

	// Gather the responses from the shards and write them to the buffer.
	if err := w.gather(ctx, diceDBCmd, len(cmdList), isWatchNotification); err != nil {
		return err
	}

	if meta.CmdType == Watch {
		// Proceed to subscribe after successful execution
		watchmanager.CmdWatchSubscriptionChan <- watchmanager.WatchSubscription{
			Subscribe:    true,
			WatchCmd:     cmdList[len(cmdList)-1],
			AdhocReqChan: w.adhocReqChan,
		}
	}

	return nil
}

// scatter distributes the DiceDB commands to the respective shards based on the key.
// For each command, it calculates the shard ID and sends the command to the shard's request channel for processing.
func (w *BaseWorker) scatter(ctx context.Context, cmds []*cmd.DiceDBCmd) error {
	// Otherwise check for the shard based on the key using hash
	// and send it to the particular shard
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for i := uint8(0); i < uint8(len(cmds)); i++ {
			var rc chan *ops.StoreOp
			var sid shard.ShardID
			var key string
			if len(cmds[i].Args) > 0 {
				key = cmds[i].Args[0]
			} else {
				key = cmds[i].Cmd
			}

			sid, rc = w.shardManager.GetShardInfo(key)

			rc <- &ops.StoreOp{
				SeqID:     i,
				RequestID: cmds[i].RequestID,
				Cmd:       cmds[i],
				WorkerID:  w.id,
				ShardID:   sid,
				Client:    nil,
			}
		}
	}

	return nil
}

// gather collects the responses from multiple shards and writes the results into the provided buffer.
// It first waits for responses from all the shards and then processes the result based on the command type (SingleShard, Custom, or Multishard).
func (w *BaseWorker) gather(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, numCmds int, isWatchNotification bool) error {
	// Loop to wait for messages from number of shards
	var evalResp []eval.EvalResponse
	for numCmds != 0 {
		select {
		case <-ctx.Done():
			w.logger.Error("Timed out waiting for response from shards", slog.String("workerID", w.id), slog.Any("error", ctx.Err()))
		case resp, ok := <-w.respChan:
			if ok {
				evalResp = append(evalResp, *resp.EvalResponse)
			}
			numCmds--
			continue
		case sError, ok := <-w.shardManager.ShardErrorChan:
			if ok {
				w.logger.Error("Error from shard", slog.String("workerID", w.id), slog.Any("error", sError))
			}
		}
	}

	val, ok := CommandsMeta[diceDBCmd.Cmd]

	if isWatchNotification {
		if evalResp[0].Error != nil {
			err := w.ioHandler.Write(ctx, querymanager.GenericWatchResponse(diceDBCmd.Cmd, fmt.Sprintf("%d", diceDBCmd.GetFingerprint()), evalResp[0].Error))
			if err != nil {
				w.logger.Debug("Error sending push response to client", slog.String("workerID", w.id), slog.Any("error", err))
			}
			return err
		}

		err := w.ioHandler.Write(ctx, querymanager.GenericWatchResponse(diceDBCmd.Cmd, fmt.Sprintf("%d", diceDBCmd.GetFingerprint()), evalResp[0].Result))
		if err != nil {
			w.logger.Debug("Error sending push response to client", slog.String("workerID", w.id), slog.Any("error", err))
			return err
		}
		return nil // Exit after handling watch case
	}

	// TODO: Remove it once we have migrated all the commands
	if !ok {
		if evalResp[0].Error != nil {
			err := w.ioHandler.Write(ctx, evalResp[0].Error)
			if err != nil {
				w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
			}
			return err
		}

		err := w.ioHandler.Write(ctx, evalResp[0].Result)
		if err != nil {
			w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
			return err
		}
	} else {
		switch val.CmdType {
		case SingleShard, Custom:
			// Handle single-shard or custom commands
			if evalResp[0].Error != nil {
				err := w.ioHandler.Write(ctx, evalResp[0].Error)
				if err != nil {
					w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
				}
				return err
			}

			err := w.ioHandler.Write(ctx, evalResp[0].Result)
			if err != nil {
				w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
				return err
			}

		case MultiShard:
			// Handle multi-shard commands
			err := w.ioHandler.Write(ctx, val.composeResponse(evalResp...))
			if err != nil {
				w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
				return err
			}

		default:
			w.logger.Error("Unknown command type", slog.String("workerID", w.id), slog.String("command", diceDBCmd.Cmd), slog.Any("evalResp", evalResp))
			err := w.ioHandler.Write(ctx, diceerrors.ErrInternalServer)
			if err != nil {
				w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
				return err
			}
		}
	}

	return nil
}

func (w *BaseWorker) isAuthenticated(diceDBCmd *cmd.DiceDBCmd) error {
	if diceDBCmd.Cmd != auth.Cmd && !w.Session.IsActive() {
		return errors.New("NOAUTH Authentication required")
	}

	return nil
}

// RespAuth returns with an encoded "OK" if the user is authenticated
// If the user is not authenticated, it returns with an encoded error message
func (w *BaseWorker) RespAuth(args []string) interface{} {
	// Check for incorrect number of arguments (arity error).
	if len(args) < 1 || len(args) > 2 {
		return diceerrors.ErrWrongArgumentCount("AUTH")
	}

	if config.DiceConfig.Auth.Password == "" {
		return diceerrors.ErrAuth
	}

	username := config.DiceConfig.Auth.UserName
	var password string

	if len(args) == 1 {
		password = args[0]
	} else {
		username, password = args[0], args[1]
	}

	if err := w.Session.Validate(username, password); err != nil {
		return err
	}

	return clientio.OK
}

func (w *BaseWorker) Stop() error {
	w.logger.Info("Stopping worker", slog.String("workerID", w.id))
	w.Session.Expire()
	return nil
}
