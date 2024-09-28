package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"syscall"
	"time"

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
	id                string
	ioHandler         iohandler.IOHandler
	parser            requestparser.Parser
	shardManager      *shard.ShardManager
	respChan          chan *ops.StoreResponse
	Session           *auth.Session
	globalErrorChan   chan error
	logger            *slog.Logger
	lastActivity      time.Time
	lastActivityMux   sync.RWMutex
	keepAliveInterval int32
	clientTimeout     int32
	commandTimeout    int32
	connectionTimeout int32
}

func NewWorker(wid string, respChan chan *ops.StoreResponse,
	ioHandler iohandler.IOHandler, parser requestparser.Parser,
	shardManager *shard.ShardManager, gec chan error,
	logger *slog.Logger) *BaseWorker {
	return &BaseWorker{
		id:                wid,
		ioHandler:         ioHandler,
		parser:            parser,
		shardManager:      shardManager,
		globalErrorChan:   gec,
		respChan:          respChan,
		logger:            logger,
		Session:           auth.NewSession(),
		lastActivity:      time.Now(),
		keepAliveInterval: config.DiceConfig.Server.KeepAlive,
		clientTimeout:     config.DiceConfig.Server.Timeout,
		connectionTimeout: config.DiceConfig.Server.Timeout,
		commandTimeout:    config.DiceConfig.Server.Timeout,
	}
}

func (w *BaseWorker) ID() string {
	return w.id
}

func (w *BaseWorker) Start(ctx context.Context) error {
	go w.keepAliveCheck(ctx)
	errChan := make(chan error, 1)
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
		default:
			clientCtx, clientCancel := context.WithTimeout(ctx, time.Duration(w.clientTimeout)*time.Second)
			data, err := w.ioHandler.Read(clientCtx)
			clientCancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					w.logger.Warn("Client read timeout", slog.String("workerID", w.id))
					continue
				}

				w.logger.Debug("Read error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
				return err
			}

			w.updateLastActivity()

			cmds, err := w.parser.Parse(data)
			if err != nil {
				err = w.ioHandler.Write(ctx, clientio.Encode(err, true))
				if err != nil {
					w.logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
			}
			if len(cmds) == 0 {
				err = w.ioHandler.Write(ctx, clientio.Encode("ERR: Invalid request", true))
				if err != nil {
					w.logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
				continue
			}

			// DiceDB supports clients to send only one request at a time
			// We also need to ensure that the client is blocked until the response is received
			if len(cmds) > 1 {
				err = w.ioHandler.Write(ctx, clientio.Encode("ERR: Multiple commands not supported", true))
				if err != nil {
					w.logger.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
			}

			err = w.isAuthenticated(cmds[0])
			if err != nil {
				werr := w.ioHandler.Write(ctx, clientio.Encode(err, false))
				if werr != nil {
					w.logger.Debug("Write error, connection closed possibly", slog.Any("error", errors.Join(err, werr)))
					return errors.Join(err, werr)
				}
			}
			// executeCommand executes the command and return the response back to the client
			go func(errChan chan<- error) {
				execctx, cancel := context.WithTimeout(ctx, 1*time.Second) // Timeout if
				defer cancel()
				err = w.executeCommand(execctx, cmds[0])
				if err != nil {
					w.logger.Error("Error executing command", slog.String("workerID", w.id), slog.Any("error", err))
					if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ETIMEDOUT) {
						w.logger.Debug("Connection closed for worker", slog.String("workerID", w.id), slog.Any("error", err))
						errChan <- err
					}
				}
			}(errChan)
		}
	}
}

func (w *BaseWorker) executeCommand(ctx context.Context, redisCmd *cmd.RedisCmd) error {
	w.updateLastActivity()

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(w.commandTimeout)*time.Second)
	defer cancel()

	resultChan := make(chan error, 1)

	go func() {
		var err error
		defer func() {
			resultChan <- err
		}()

	// Break down the single command into multiple commands if multisharding is supported.
	// The length of cmdList helps determine how many shards to wait for responses.
	cmdList := make([]*cmd.RedisCmd, 0)

	// Retrieve metadata for the command to determine if multisharding is supported.
	meta, ok := CommandsMeta[redisCmd.Cmd]
	if !ok {
		// If no metadata exists, treat it as a single command.
		cmdList = append(cmdList, redisCmd)
	} else {
		// Depending on the command type, decide how to handle it.
		switch meta.CmdType {
		case Global:
			// If it's a global command, process it immediately without involving any shards.
			err := w.ioHandler.Write(ctx, meta.WorkerCommandHandler(redisCmd.Args))
			w.logger.Debug("Error executing for worker", slog.String("workerID", w.id), slog.Any("error", err))
			return err

			case SingleShard:
				// For single-shard or custom commands, process them without breaking up.
				cmdList = append(cmdList, redisCmd)

			case MultiShard:
				// If the command supports multisharding, break it down into multiple commands.
				cmdList = meta.decomposeCommand(redisCmd)
			case Custom:
				switch redisCmd.Cmd {
				case CmdAuth:
					err := w.ioHandler.Write(cmdCtx, w.RespAuth(redisCmd.Args))
					w.logger.Error("Error sending auth response to worker", slog.String("workerID", w.id), slog.Any("error", err))

				case CmdAbort:
					w.logger.Info("Received ABORT command, initiating server shutdown", slog.String("workerID", w.id))
					w.globalErrorChan <- diceerrors.ErrAborted
					err = nil
				default:
					cmdList = append(cmdList, redisCmd)
				}
			}
		}

		// Scatter the broken-down commands to the appropriate shards.
		err = w.scatter(cmdCtx, cmdList)
		if err != nil {
			return
		}

		// Gather the responses from the shards and write them to the buffer.
		err = w.gather(cmdCtx, redisCmd.Cmd, len(cmdList), meta.CmdType)
		if err != nil {
			return
		}
	}()

	select {
	case <-cmdCtx.Done():
		if cmdCtx.Err() == context.DeadlineExceeded {
			w.logger.Warn("Command execution timed out", slog.String("workerID", w.id), slog.String("command", redisCmd.Cmd))
			return fmt.Errorf("command execution timed out: %w", cmdCtx.Err())
		}
		return cmdCtx.Err()
	case err := <-resultChan:
		return err
	}
}

// scatter distributes the Redis commands to the respective shards based on the key.
// For each command, it calculates the shard ID and sends the command to the shard's request channel for processing.
func (w *BaseWorker) scatter(ctx context.Context, cmds []*cmd.RedisCmd) error {
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
func (w *BaseWorker) gather(ctx context.Context, c string, numCmds int, ct CmdType) error {
	// Loop to wait for messages from numberof shards
	var evalResp []eval.EvalResponse
	for numCmds != 0 {
		select {
		case <-ctx.Done():
			w.logger.Error("Timed out waiting for response from shards", slog.String("workerID", w.id), slog.Any("error", ctx.Err()))
		case resp, ok := <-w.respChan:
			if ok {
				evalResp = append(evalResp, resp.EvalResponse)
			}
			numCmds--
			continue
		case sError, ok := <-w.shardManager.ShardErrorChan:
			if ok {
				w.logger.Error("Error from shard", slog.String("workerID", w.id), slog.Any("error", sError))
			}
		}
	}

	// TODO: This is a temporary solution. In the future, all commands should be refactored to be multi-shard compatible.
	// TODO: There are a few commands such as QWATCH, RENAME, MGET, MSET that wouldn't work in multi-shard mode without refactoring.
	// TODO: These commands should be refactored to be multi-shard compatible before DICE-DB is completely multi-shard.
	// Check if command is part of the new WorkerCommandsMeta map i.e. if the command has been refactored to be multi-shard compatible.
	// If not found, treat it as a command that's not yet refactored, and write the response back to the client.
	val, ok := CommandsMeta[c]
	if !ok {
		if evalResp[0].Error != nil {
			err := w.ioHandler.Write(ctx, []byte(evalResp[0].Error.Error()))
			if err != nil {
				w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
				return err
			}
		}

		err := w.ioHandler.Write(ctx, evalResp[0].Result.([]byte))
		if err != nil {
			w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
			return err
		}

		return nil
	}

	switch ct {
	case SingleShard, Custom:
		if evalResp[0].Error != nil {
			err := w.ioHandler.Write(ctx, []byte(evalResp[0].Error.Error()))
			if err != nil {
				w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
			}

			return err
		}

		err := w.ioHandler.Write(ctx, evalResp[0].Result.([]byte))
		if err != nil {
			w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
			return err
		}

	case MultiShard:
		err := w.ioHandler.Write(ctx, val.composeResponse(evalResp...))
		if err != nil {
			w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
			return err
		}

	default:
		w.logger.Error("Unknown command type", slog.String("workerID", w.id))
		err := w.ioHandler.Write(ctx, []byte(diceerrors.InternalServerError))
		if err != nil {
			w.logger.Debug("Error sending response to client", slog.String("workerID", w.id), slog.Any("error", err))
			return err
		}
	}

	return nil
}

func (w *BaseWorker) isAuthenticated(redisCmd *cmd.RedisCmd) error {
	if redisCmd.Cmd != auth.Cmd && !w.Session.IsActive() {
		return errors.New("NOAUTH Authentication required")
	}

	return nil
}

// RespAuth returns with an encoded "OK" if the user is authenticated
// If the user is not authenticated, it returns with an encoded error message
func (w *BaseWorker) RespAuth(args []string) []byte {
	// Check for incorrect number of arguments (arity error).
	if len(args) < 1 || len(args) > 2 {
		return diceerrors.NewErrArity("AUTH") // Return an error if the number of arguments is not equal to 1.
	}

	if config.DiceConfig.Auth.Password == "" {
		return diceerrors.NewErrWithMessage("AUTH <password> called without any password configured for the default user. Are you sure your configuration is correct?")
	}

	username := config.DiceConfig.Auth.UserName
	var password string

	if len(args) == 1 {
		password = args[0]
	} else {
		username, password = args[0], args[1]
	}

	if err := w.Session.Validate(username, password); err != nil {
		return clientio.Encode(err, false)
	}

	return clientio.RespOK
}

func (w *BaseWorker) Stop() error {
	w.logger.Info("Stopping worker", slog.String("workerID", w.id))
	w.Session.Expire()
	return nil
}

func (w *BaseWorker) updateLastActivity() {
	w.lastActivityMux.Lock()
	w.lastActivity = time.Now()
	w.lastActivityMux.Unlock()
}

func (w *BaseWorker) keepAliveCheck(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.keepAliveInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping keepAliveCheck due to context cancellation", slog.String("workerID", w.id))
			return
		case <-ticker.C:
			w.lastActivityMux.RLock()
			lastActivity := w.lastActivity
			w.lastActivityMux.RUnlock()

			if time.Since(lastActivity) > time.Duration(w.keepAliveInterval)*time.Second {
				w.logger.Warn("Connection timeout for worker", slog.String("workerID", w.id))
				err := w.Stop()
				if err != nil {
					w.logger.Error("Error stopping worker after timeout", slog.String("workerID", w.id), slog.Any("error", err))
				}
				return
			}
		}
	}
}
