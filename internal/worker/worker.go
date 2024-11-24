package worker

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/clientio/iohandler"
	"github.com/dicedb/dice/internal/clientio/requestparser"
	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/querymanager"
	"github.com/dicedb/dice/internal/shard"
	"github.com/dicedb/dice/internal/wal"
	"github.com/dicedb/dice/internal/watchmanager"
	"github.com/google/uuid"
)

var (
	requestCounter uint32
)

// Worker interface
type Worker interface {
	ID() string
	Start(context.Context) error
	Stop() error
}

type BaseWorker struct {
	id                       string
	ioHandler                iohandler.IOHandler
	parser                   requestparser.Parser
	shardManager             *shard.ShardManager
	adhocReqChan             chan *cmd.DiceDBCmd
	Session                  *auth.Session
	globalErrorChan          chan error
	responseChan             chan *ops.StoreResponse
	preprocessingChan        chan *ops.StoreResponse
	cmdWatchSubscriptionChan chan watchmanager.WatchSubscription
	wl                       wal.AbstractWAL
}

func NewWorker(wid string, responseChan, preprocessingChan chan *ops.StoreResponse,
	cmdWatchSubscriptionChan chan watchmanager.WatchSubscription,
	ioHandler iohandler.IOHandler, parser requestparser.Parser,
	shardManager *shard.ShardManager, gec chan error, wl wal.AbstractWAL) *BaseWorker {
	return &BaseWorker{
		id:                       wid,
		ioHandler:                ioHandler,
		parser:                   parser,
		shardManager:             shardManager,
		globalErrorChan:          gec,
		responseChan:             responseChan,
		preprocessingChan:        preprocessingChan,
		Session:                  auth.NewSession(),
		adhocReqChan:             make(chan *cmd.DiceDBCmd, config.DiceConfig.Performance.AdhocReqChanBufSize),
		cmdWatchSubscriptionChan: cmdWatchSubscriptionChan,
		wl:                       wl,
	}
}

func (w *BaseWorker) ID() string {
	return w.id
}

func (w *BaseWorker) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	dataChan := make(chan []byte)
	readErrChan := make(chan error)

	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	go func() {
		defer close(dataChan)
		defer close(readErrChan)

		for {
			data, err := w.ioHandler.Read(runCtx)
			if err != nil {
				select {
				case readErrChan <- err:
				case <-runCtx.Done(): // exit if worker exits
				}
				return
			}

			select {
			case dataChan <- data:
			case <-runCtx.Done(): // exit if worker exits
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			err := w.Stop()
			if err != nil {
				slog.Warn("Error stopping worker:", slog.String("workerID", w.id), slog.Any("error", err))
			}
			return ctx.Err()
		case err := <-errChan:
			if err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
					slog.Debug("Connection closed for worker", slog.String("workerID", w.id), slog.Any("error", err))
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
					slog.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
			}
			if len(cmds) == 0 {
				err = w.ioHandler.Write(ctx, fmt.Errorf("ERR: Invalid request"))
				if err != nil {
					slog.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
				continue
			}

			// DiceDB supports clients to send only one request at a time
			// We also need to ensure that the client is blocked until the response is received
			if len(cmds) > 1 {
				err = w.ioHandler.Write(ctx, fmt.Errorf("ERR: Multiple commands not supported"))
				if err != nil {
					slog.Debug("Write error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
					return err
				}
			}

			err = w.isAuthenticated(cmds[0])
			if err != nil {
				werr := w.ioHandler.Write(ctx, err)
				if werr != nil {
					slog.Debug("Write error, connection closed possibly", slog.Any("error", errors.Join(err, werr)))
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
			slog.Debug("Read error, connection closed possibly", slog.String("workerID", w.id), slog.Any("error", err))
			return err
		}
	}
}

func (w *BaseWorker) executeCommandHandler(execCtx context.Context, errChan chan error, cmds []*cmd.DiceDBCmd, isWatchNotification bool) {
	// Retrieve metadata for the command to determine if multisharding is supported.
	meta, ok := CommandsMeta[cmds[0].Cmd]
	if ok && meta.preProcessing {
		if err := meta.preProcessResponse(w, cmds[0]); err != nil {
			e := w.ioHandler.Write(execCtx, err)
			if e != nil {
				slog.Debug("Error executing for worker", slog.String("workerID", w.id), slog.Any("error", err))
			}
		}
	}

	err := w.executeCommand(execCtx, cmds[0], isWatchNotification)
	if err != nil {
		slog.Error("Error executing command", slog.String("workerID", w.id), slog.Any("error", err))
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ETIMEDOUT) {
			slog.Debug("Connection closed for worker", slog.String("workerID", w.id), slog.Any("error", err))
			errChan <- err
		}
	}
}

func (w *BaseWorker) executeCommand(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, isWatchNotification bool) error {
	// Break down the single command into multiple commands if multisharding is supported.
	// The length of cmdList helps determine how many shards to wait for responses.
	cmdList := make([]*cmd.DiceDBCmd, 0)
	var watchLabel string

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
			slog.Debug("Error executing for worker", slog.String("workerID", w.id), slog.Any("error", err))
			return err

		case SingleShard:
			// For single-shard or custom commands, process them without breaking up.
			cmdList = append(cmdList, diceDBCmd)

		case MultiShard, AllShard:
			var err error
			// If the command supports multisharding, break it down into multiple commands.
			cmdList, err = meta.decomposeCommand(ctx, w, diceDBCmd)
			if err != nil {
				var workerErr error
				// Check if it's a CustomError
				var customErr *diceerrors.PreProcessError
				if errors.As(err, &customErr) {
					workerErr = w.ioHandler.Write(ctx, customErr.Result)
				} else {
					workerErr = w.ioHandler.Write(ctx, err)
				}
				if workerErr != nil {
					slog.Debug("Error executing for worker", slog.String("workerID", w.id), slog.Any("error", workerErr))
				}
				return workerErr
			}

		case Custom:
			return w.handleCustomCommands(ctx, diceDBCmd)

		case Watch:
			// Generate the Cmd being watched. All we need to do is remove the .WATCH suffix from the command and pass
			// it along as is.
			// Modify the command name to remove the .WATCH suffix, this will allow us to generate a consistent
			// fingerprint (which uses the command name without the suffix)
			diceDBCmd.Cmd = diceDBCmd.Cmd[:len(diceDBCmd.Cmd)-6]

			// check if the last argument is a watch label
			label := diceDBCmd.Args[len(diceDBCmd.Args)-1]
			if _, err := uuid.Parse(label); err == nil {
				watchLabel = label

				// remove the watch label from the args
				diceDBCmd.Args = diceDBCmd.Args[:len(diceDBCmd.Args)-1]
			}

			watchCmd := &cmd.DiceDBCmd{
				Cmd:  diceDBCmd.Cmd,
				Args: diceDBCmd.Args,
			}
			cmdList = append(cmdList, watchCmd)
			isWatchNotification = true

		case Unwatch:
			// Generate the Cmd being unwatched. All we need to do is remove the .UNWATCH suffix from the command and pass
			// it along as is.
			// Modify the command name to remove the .UNWATCH suffix, this will allow us to generate a consistent
			// fingerprint (which uses the command name without the suffix)
			diceDBCmd.Cmd = diceDBCmd.Cmd[:len(diceDBCmd.Cmd)-8]
			watchCmd := &cmd.DiceDBCmd{
				Cmd:  diceDBCmd.Cmd,
				Args: diceDBCmd.Args,
			}
			cmdList = append(cmdList, watchCmd)
			isWatchNotification = false
		}
	}

	// Unsubscribe Unwatch command type
	if meta.CmdType == Unwatch {
		return w.handleCommandUnwatch(ctx, cmdList)
	}

	// Scatter the broken-down commands to the appropriate shards.
	if err := w.scatter(ctx, cmdList, meta.CmdType); err != nil {
		return err
	}

	// Gather the responses from the shards and write them to the buffer.
	if err := w.gather(ctx, diceDBCmd, len(cmdList), isWatchNotification, watchLabel); err != nil {
		return err
	}

	if meta.CmdType == Watch {
		// Proceed to subscribe after successful execution
		w.handleCommandWatch(cmdList)
	}

	return nil
}

func (w *BaseWorker) handleCustomCommands(ctx context.Context, diceDBCmd *cmd.DiceDBCmd) error {
	// if command is of type Custom, write a custom logic around it
	switch diceDBCmd.Cmd {
	case CmdAuth:
		err := w.ioHandler.Write(ctx, w.RespAuth(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending auth response to worker", slog.String("workerID", w.id), slog.Any("error", err))
		}
		return err
	case CmdEcho:
		err := w.ioHandler.Write(ctx, RespEcho(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending echo response to worker", slog.String("workerID", w.id), slog.Any("error", err))
		}
		return err
	case CmdAbort:
		err := w.ioHandler.Write(ctx, clientio.OK)
		if err != nil {
			slog.Error("Error sending abort response to worker", slog.String("workerID", w.id), slog.Any("error", err))
		}
		slog.Info("Received ABORT command, initiating server shutdown", slog.String("workerID", w.id))
		w.globalErrorChan <- diceerrors.ErrAborted
		return err
	case CmdPing:
		err := w.ioHandler.Write(ctx, RespPING(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending ping response to worker", slog.String("workerID", w.id), slog.Any("error", err))
		}
		return err
	case CmdHello:
		err := w.ioHandler.Write(ctx, RespHello(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending ping response to worker", slog.String("workerID", w.id), slog.Any("error", err))
		}
		return err
	case CmdSleep:
		err := w.ioHandler.Write(ctx, RespSleep(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending ping response to worker", slog.String("workerID", w.id), slog.Any("error", err))
		}
		return err
	default:
		return diceerrors.ErrUnknownCmd(diceDBCmd.Cmd)
	}
}

// handleCommandWatch sends a watch subscription request to the watch manager.
func (w *BaseWorker) handleCommandWatch(cmdList []*cmd.DiceDBCmd) {
	w.cmdWatchSubscriptionChan <- watchmanager.WatchSubscription{
		Subscribe:    true,
		WatchCmd:     cmdList[len(cmdList)-1],
		AdhocReqChan: w.adhocReqChan,
	}
}

// handleCommandUnwatch sends an unwatch subscription request to the watch manager. It also sends a response to the client.
// The response is sent before the unwatch request is processed by the watch manager.
func (w *BaseWorker) handleCommandUnwatch(ctx context.Context, cmdList []*cmd.DiceDBCmd) error {
	// extract the fingerprint
	command := cmdList[len(cmdList)-1]
	fp, parseErr := strconv.ParseUint(command.Args[0], 10, 32)
	if parseErr != nil {
		err := w.ioHandler.Write(ctx, diceerrors.ErrInvalidFingerprint)
		if err != nil {
			return fmt.Errorf("error sending push response to client: %v", err)
		}
		return parseErr
	}

	// send the unsubscribe request
	w.cmdWatchSubscriptionChan <- watchmanager.WatchSubscription{
		Subscribe:    false,
		AdhocReqChan: w.adhocReqChan,
		Fingerprint:  uint32(fp),
	}

	err := w.ioHandler.Write(ctx, clientio.RespOK)
	if err != nil {
		return fmt.Errorf("error sending push response to client: %v", err)
	}
	return nil
}

// scatter distributes the DiceDB commands to the respective shards based on the key.
// For each command, it calculates the shard ID and sends the command to the shard's request channel for processing.
func (w *BaseWorker) scatter(ctx context.Context, cmds []*cmd.DiceDBCmd, cmdType CmdType) error {
	// Otherwise check for the shard based on the key using hash
	// and send it to the particular shard
	// Check if the context has been canceled or expired.
	select {
	case <-ctx.Done():
		// If the context is canceled, return the error associated with it.
		return ctx.Err()
	default:
		// Proceed with the default case when the context is not canceled.

		if cmdType == AllShard {
			// If the command type is for all shards, iterate over all available shards.
			for i := uint8(0); i < uint8(w.shardManager.GetShardCount()); i++ {
				// Get the shard ID (i) and its associated request channel.
				shardID, responseChan := i, w.shardManager.GetShard(i).ReqChan

				// Send a StoreOp operation to the shard's request channel.
				responseChan <- &ops.StoreOp{
					SeqID:     i,                         // Sequence ID for this operation.
					RequestID: GenerateUniqueRequestID(), // Unique identifier for the request.
					Cmd:       cmds[0],                   // Command to be executed, using the first command in cmds.
					WorkerID:  w.id,                      // ID of the current worker.
					ShardID:   shardID,                   // ID of the shard handling this operation.
					Client:    nil,                       // Client information (if applicable).
				}
			}
		} else {
			// If the command type is specific to certain commands, process them individually.
			for i := uint8(0); i < uint8(len(cmds)); i++ {
				// Determine the appropriate shard for the current command using a routing key.
				shardID, responseChan := w.shardManager.GetShardInfo(getRoutingKeyFromCommand(cmds[i]))

				// Send a StoreOp operation to the shard's request channel.
				responseChan <- &ops.StoreOp{
					SeqID:     i,                         // Sequence ID for this operation.
					RequestID: GenerateUniqueRequestID(), // Unique identifier for the request.
					Cmd:       cmds[i],                   // Command to be executed, using the current command in cmds.
					WorkerID:  w.id,                      // ID of the current worker.
					ShardID:   shardID,                   // ID of the shard handling this operation.
					Client:    nil,                       // Client information (if applicable).
				}
			}
		}
	}

	return nil
}

// getRoutingKeyFromCommand determines the key used for shard routing
func getRoutingKeyFromCommand(diceDBCmd *cmd.DiceDBCmd) string {
	if len(diceDBCmd.Args) > 0 {
		return diceDBCmd.Args[0]
	}
	return diceDBCmd.Cmd
}

// gather collects the responses from multiple shards and writes the results into the provided buffer.
// It first waits for responses from all the shards and then processes the result based on the command type (SingleShard, Custom, or Multishard).
func (w *BaseWorker) gather(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, numCmds int, isWatchNotification bool, watchLabel string) error {
	// Collect responses from all shards
	storeOp, err := w.gatherResponses(ctx, numCmds)
	if err != nil {
		return err
	}

	if len(storeOp) == 0 {
		slog.Error("No response from shards",
			slog.String("workerID", w.id),
			slog.String("command", diceDBCmd.Cmd))
		return fmt.Errorf("no response from shards for command: %s", diceDBCmd.Cmd)
	}

	if isWatchNotification {
		return w.handleWatchNotification(ctx, diceDBCmd, storeOp[0], watchLabel)
	}

	// Process command based on its type
	cmdMeta, ok := CommandsMeta[diceDBCmd.Cmd]
	if !ok {
		return w.handleLegacyCommand(ctx, storeOp[0])
	}

	return w.handleCommand(ctx, cmdMeta, diceDBCmd, storeOp)
}

// gatherResponses collects responses from all shards
func (w *BaseWorker) gatherResponses(ctx context.Context, numCmds int) ([]ops.StoreResponse, error) {
	storeOp := make([]ops.StoreResponse, 0, numCmds)

	for numCmds > 0 {
		select {
		case <-ctx.Done():
			slog.Error("Timed out waiting for response from shards",
				slog.String("workerID", w.id),
				slog.Any("error", ctx.Err()))
			return nil, ctx.Err()

		case resp, ok := <-w.responseChan:
			if ok {
				storeOp = append(storeOp, *resp)
			}
			numCmds--

		case sError, ok := <-w.shardManager.ShardErrorChan:
			if ok {
				slog.Error("Error from shard",
					slog.String("workerID", w.id),
					slog.Any("error", sError))
				return nil, sError.Error
			}
		}
	}

	return storeOp, nil
}

// handleWatchNotification processes watch notification responses
func (w *BaseWorker) handleWatchNotification(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, resp ops.StoreResponse, watchLabel string) error {
	fingerprint := fmt.Sprintf("%d", diceDBCmd.GetFingerprint())

	// if watch label is not empty, then this is the first response for the watch command
	// hence, we will send the watch label as part of the response
	firstRespElem := diceDBCmd.Cmd
	if watchLabel != "" {
		firstRespElem = watchLabel
	}

	if resp.EvalResponse.Error != nil {
		return w.writeResponse(ctx, querymanager.GenericWatchResponse(firstRespElem, fingerprint, resp.EvalResponse.Error))
	}

	return w.writeResponse(ctx, querymanager.GenericWatchResponse(firstRespElem, fingerprint, resp.EvalResponse.Result))
}

// handleLegacyCommand processes commands not in CommandsMeta
func (w *BaseWorker) handleLegacyCommand(ctx context.Context, resp ops.StoreResponse) error {
	if resp.EvalResponse.Error != nil {
		return w.writeResponse(ctx, resp.EvalResponse.Error)
	}
	return w.writeResponse(ctx, resp.EvalResponse.Result)
}

// handleCommand processes commands based on their type
func (w *BaseWorker) handleCommand(ctx context.Context, cmdMeta CmdMeta, diceDBCmd *cmd.DiceDBCmd, storeOp []ops.StoreResponse) error {
	var err error

	switch cmdMeta.CmdType {
	case SingleShard, Custom:
		if storeOp[0].EvalResponse.Error != nil {
			err = w.writeResponse(ctx, storeOp[0].EvalResponse.Error)
		} else {
			err = w.writeResponse(ctx, storeOp[0].EvalResponse.Result)
		}

		if err == nil && w.wl != nil {
			w.wl.LogCommand(diceDBCmd)
		}
	case MultiShard, AllShard:
		err = w.writeResponse(ctx, cmdMeta.composeResponse(storeOp...))

		if err == nil && w.wl != nil {
			w.wl.LogCommand(diceDBCmd)
		}
	default:
		slog.Error("Unknown command type",
			slog.String("workerID", w.id),
			slog.String("command", diceDBCmd.Cmd),
			slog.Any("evalResp", storeOp))
		err = w.writeResponse(ctx, diceerrors.ErrInternalServer)
	}
	return err
}

// writeResponse handles writing responses and logging errors
func (w *BaseWorker) writeResponse(ctx context.Context, response interface{}) error {
	err := w.ioHandler.Write(ctx, response)
	if err != nil {
		slog.Debug("Error sending response to client",
			slog.String("workerID", w.id),
			slog.Any("error", err))
	}
	return err
}

func (w *BaseWorker) isAuthenticated(diceDBCmd *cmd.DiceDBCmd) error {
	if diceDBCmd.Cmd != auth.Cmd && !w.Session.IsActive() {
		return errors.New("NOAUTH Authentication required")
	}

	return nil
}

func (w *BaseWorker) Stop() error {
	slog.Info("Stopping worker", slog.String("workerID", w.id))
	w.Session.Expire()
	return nil
}

func GenerateRandomUint32() (uint32, error) {
	var b [4]byte             // Create a byte array to hold the random bytes
	_, err := rand.Read(b[:]) // Fill the byte array with secure random bytes
	if err != nil {
		return 0, err // Return an error if reading failed
	}
	return binary.BigEndian.Uint32(b[:]), nil // Convert bytes to uint32
}

func GenerateUniqueRequestID() uint32 {
	return atomic.AddUint32(&requestCounter, 1)
}
