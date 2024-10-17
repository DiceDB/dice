package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/comm"
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

type CommandHandler struct {
	ShardManager       *shard.ShardManager
	RespChan           chan *ops.StoreResponse
	GlobalErrorChannel chan error
	Logger             *slog.Logger
	ID                 string
}

type CommandResponse struct {
	ResponseData interface{}
	Error        error
	Action       string
	Args         []string
}

func (ch *CommandHandler) ExecuteCommand(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, isHTTPOp, isWebsocketOp bool, client *comm.Client) *CommandResponse {
	cmdList := make([]*cmd.DiceDBCmd, 0)
	meta, ok := CommandsMeta[diceDBCmd.Cmd]
	if !ok {
		cmdList = append(cmdList, diceDBCmd)
	} else {
		switch meta.CmdType {
		case Global:
			return ch.handleGlobalCommand(diceDBCmd)
		case Custom:
			return ch.handleCustomCommand(diceDBCmd)
		case SingleShard:
			cmdList = append(cmdList, diceDBCmd)
		case MultiShard:
			cmdList = meta.decomposeCommand(diceDBCmd)
		case Watch:
			modifiedCmd := diceDBCmd.Cmd[:len(diceDBCmd.Cmd)-6]
			watchCmd := &cmd.DiceDBCmd{
				Cmd:  modifiedCmd,
				Args: diceDBCmd.Args,
			}
			cmdList = append(cmdList, watchCmd)
		}
	}

	err := ch.scatter(ctx, cmdList, isHTTPOp, isWebsocketOp, client)
	if err != nil {
		return &CommandResponse{Error: err}
	}

	responseData, err := ch.gather(ctx, diceDBCmd.Cmd, len(cmdList), meta.CmdType)
	if err != nil {
		return &CommandResponse{Error: err}
	}

	return &CommandResponse{ResponseData: responseData}
}

func (ch *CommandHandler) handleGlobalCommand(redisCmd *cmd.DiceDBCmd) *CommandResponse {
	return &CommandResponse{Args: redisCmd.Args, Action: CmdPing}
}

func (ch *CommandHandler) handleCustomCommand(redisCmd *cmd.DiceDBCmd) *CommandResponse {
	switch redisCmd.Cmd {
	case CmdAuth:
		return &CommandResponse{Args: redisCmd.Args, Action: CmdAuth}
	case CmdAbort:
		return &CommandResponse{Action: CmdAbort}
	default:
		return &CommandResponse{Error: fmt.Errorf("unknown custom command: %s", redisCmd.Cmd)}
	}
}

func (ch *CommandHandler) scatter(ctx context.Context, cmds []*cmd.DiceDBCmd, isHTTPOp, isWebsocketOp bool, client *comm.Client) error {
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

			sid, rc = ch.ShardManager.GetShardInfo(key)
			if rc == nil {
				ch.Logger.Error("Shard channel is nil", slog.String("workerID", ch.ID), slog.Any("key", key))
				return fmt.Errorf("shard channel is nil for key: %s", key)
			}

			rc <- &ops.StoreOp{
				SeqID:       i,
				RequestID:   cmds[i].RequestID,
				Cmd:         cmds[i],
				WorkerID:    ch.ID,
				ShardID:     sid,
				Client:      client,
				HTTPOp:      isHTTPOp,
				WebsocketOp: isWebsocketOp,
			}
		}
	}

	return nil
}

func (ch *CommandHandler) gather(ctx context.Context, c string, numCmds int, ct CmdType) (interface{}, error) {
	var evalResp []eval.EvalResponse
	for numCmds != 0 {
		select {
		case <-ctx.Done():
			ch.Logger.Error("Timed out waiting for response from shards", slog.String("workerID", ch.ID), slog.Any("error", ctx.Err()))
			return nil, ctx.Err()
		case resp, ok := <-ch.RespChan:
			if ok {
				evalResp = append(evalResp, *resp.EvalResponse)
			} else {
				ch.Logger.Warn("Response channel closed", slog.String("workerID", ch.ID))
				numCmds--
			}
			numCmds--
		case sError, ok := <-ch.ShardManager.ShardErrorChan:
			if ok {
				ch.Logger.Error("Error from shard", slog.String("workerID", ch.ID), slog.Any("error", sError))
			}
		}
	}

	// Process the responses based on the command type
	val, ok := CommandsMeta[c]

	if !ok {
		// If the command is not in CommandsMeta, handle it as a default case
		if len(evalResp) == 0 {
			return nil, fmt.Errorf("no response from shards for command: %s", c)
		}
		if evalResp[0].Error != nil {
			return nil, evalResp[0].Error
		}

		return evalResp[0].Result, nil
	}

	// Handle based on command type
	switch ct {
	case SingleShard, Custom, Watch:
		if len(evalResp) == 0 {
			return nil, fmt.Errorf("no response from shards for command: %s", c)
		}
		if evalResp[0].Error != nil {
			return nil, evalResp[0].Error
		}

		return evalResp[0].Result, nil
	case MultiShard:
		// For MultiShard commands, compose the response from multiple shard responses
		responseData := val.composeResponse(evalResp...)
		return responseData, nil
	default:
		ch.Logger.Error("Unknown command type", slog.String("workerID", ch.ID), slog.String("command", c))
		return nil, diceerrors.ErrInternalServer
	}
}

type BaseWorker struct {
	ch           *CommandHandler
	ioHandler    iohandler.IOHandler
	parser       requestparser.Parser
	adhocReqChan chan *cmd.DiceDBCmd
	Session      *auth.Session
}

func NewWorker(wid string, respChan chan *ops.StoreResponse,
	ioHandler iohandler.IOHandler, parser requestparser.Parser,
	shardManager *shard.ShardManager, gec chan error,
	logger *slog.Logger) *BaseWorker {
	return &BaseWorker{
		ch: &CommandHandler{
			ShardManager:       shardManager,
			ID:                 wid,
			RespChan:           respChan,
			Logger:             logger,
			GlobalErrorChannel: gec,
		},
		ioHandler:    ioHandler,
		parser:       parser,
		Session:      auth.NewSession(),
		adhocReqChan: make(chan *cmd.DiceDBCmd, config.DiceConfig.Performance.AdhocReqChanBufSize),
	}
}

func (w *BaseWorker) ID() string {
	return w.ch.ID
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
				w.ch.Logger.Warn("Error stopping worker:", slog.String("workerID", w.ch.ID), slog.Any("error", err))
			}
			return ctx.Err()
		case err := <-errChan:
			if err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
					w.ch.Logger.Error("Connection closed for worker", slog.String("workerID", w.ch.ID), slog.Any("error", err))
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
					w.ch.Logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.ch.ID), slog.Any("error", err))
					return err
				}
			}
			if len(cmds) == 0 {
				err = w.ioHandler.Write(ctx, fmt.Errorf("ERR: Invalid request"))
				if err != nil {
					w.ch.Logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.ch.ID), slog.Any("error", err))
					return err
				}
				continue
			}

			// DiceDB supports clients to send only one request at a time
			// We also need to ensure that the client is blocked until the response is received
			if len(cmds) > 1 {
				err = w.ioHandler.Write(ctx, fmt.Errorf("ERR: Multiple commands not supported"))
				if err != nil {
					w.ch.Logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.ch.ID), slog.Any("error", err))
					return err
				}
			}

			err = w.isAuthenticated(cmds[0])
			if err != nil {
				werr := w.ioHandler.Write(ctx, err)
				if werr != nil {
					w.ch.Logger.Debug("Write error, connection closed possibly", slog.Any("error", errors.Join(err, werr)))
					return errors.Join(err, werr)
				}
			}
			// executeCommand executes the command and return the response back to the client
			func(errChan chan error) {
				execctx, cancel := context.WithTimeout(ctx, 6*time.Second) // Timeout set to 6 seconds for integration tests
				defer cancel()
				w.executeCommandHandler(execctx, errChan, cmds, false)
			}(errChan)
		}
	}
}

func (w *BaseWorker) executeCommandHandler(execCtx context.Context, errChan chan error, cmds []*cmd.DiceDBCmd, isWatchNotification bool) {
	err := w.executeCommand(execCtx, cmds[0], isWatchNotification)
	if err != nil {
		w.ch.Logger.Error("Error executing command", slog.String("workerID", w.ch.ID), slog.Any("error", err))
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ETIMEDOUT) {
			w.ch.Logger.Debug("Connection closed for worker", slog.String("workerID", w.ch.ID), slog.Any("error", err))
			errChan <- err
		}
	}
}

func (w *BaseWorker) executeCommand(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, isWatchNotification bool) error {
	// Break down the single command into multiple commands if multisharding is supported.
	// The length of cmdList helps determine how many shards to wait for responses.
	result := w.ch.ExecuteCommand(ctx, diceDBCmd, false, false, nil)

	meta := CommandsMeta[diceDBCmd.Cmd]
	if meta.CmdType == Watch || isWatchNotification {
		if result.Error != nil {
			err := w.ioHandler.Write(ctx, querymanager.GenericWatchResponse(diceDBCmd.Cmd, fmt.Sprintf("%d", diceDBCmd.GetFingerprint()), result.Error))
			if err != nil {
				w.ch.Logger.Debug("Error sending push response to client", slog.String("workerID", w.ch.ID), slog.Any("error", err))
			}
			return err
		}

		err := w.ioHandler.Write(ctx, querymanager.GenericWatchResponse(diceDBCmd.Cmd, fmt.Sprintf("%d", diceDBCmd.GetFingerprint()), result.ResponseData))
		if err != nil {
			w.ch.Logger.Debug("Error sending push response to client", slog.String("workerID", w.ch.ID), slog.Any("error", err))
			return err
		}

		if !isWatchNotification {
			diceDBCmd.Cmd = diceDBCmd.Cmd[:len(diceDBCmd.Cmd)-6]
			watchmanager.CmdWatchSubscriptionChan <- watchmanager.WatchSubscription{
				Subscribe:    true,
				WatchCmd:     diceDBCmd,
				AdhocReqChan: w.adhocReqChan,
			}
		}

		return nil // Exit after handling watch case
	}
	// Write the result back to the client
	if result.Error != nil {
		err := w.ioHandler.Write(ctx, result.Error)
		if err != nil {
			w.ch.Logger.Debug("Error sending response to client", slog.String("workerID", w.ch.ID), slog.Any("error", err))
			return err
		}
	}

	switch result.Action {
	case CmdPing:
		err := w.ioHandler.Write(ctx, meta.WorkerCommandHandler(diceDBCmd.Args))
		return err
	case CmdAbort:
		err := w.ioHandler.Write(ctx, clientio.OK)
		if err != nil {
			w.ch.Logger.Error("Error sending abort response to worker", slog.String("workerID", w.ch.ID), slog.Any("error", err))
		}
		w.ch.Logger.Info("Received ABORT command, initiating server shutdown", slog.String("workerID", w.ch.ID))
		w.ch.GlobalErrorChannel <- diceerrors.ErrAborted
		return err
	case CmdAuth:
		err := w.ioHandler.Write(ctx, w.RespAuth(diceDBCmd.Args))
		return err
	}

	err := w.ioHandler.Write(ctx, result.ResponseData)
	if err != nil {
		w.ch.Logger.Debug("Error sending response to client", slog.String("workerID", w.ch.ID), slog.Any("error", err))
		return err
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
	w.ch.Logger.Info("Stopping worker", slog.String("workerID", w.ch.ID))
	w.Session.Expire()
	return nil
}
