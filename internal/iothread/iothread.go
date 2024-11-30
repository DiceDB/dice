package iothread

import (
	"context"
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

const defaultRequestTimeout = 6 * time.Second

var requestCounter uint32

// IOThread interface
type IOThread interface {
	ID() string
	Start(context.Context) error
	Stop() error
}

type BaseIOThread struct {
	IOThread
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

func NewIOThread(wid string, responseChan, preprocessingChan chan *ops.StoreResponse,
	cmdWatchSubscriptionChan chan watchmanager.WatchSubscription,
	ioHandler iohandler.IOHandler, parser requestparser.Parser,
	shardManager *shard.ShardManager, gec chan error, wl wal.AbstractWAL) *BaseIOThread {
	return &BaseIOThread{
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

func (t *BaseIOThread) ID() string {
	return t.id
}

func (t *BaseIOThread) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	incomingDataChan := make(chan []byte)
	readErrChan := make(chan error)

	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	// This method is run in a separate goroutine to ensure that the main event loop in the Start method
	// remains non-blocking and responsive to other events, such as adhoc requests or context cancellations.
	go t.startInputReader(runCtx, incomingDataChan, readErrChan)

	for {
		select {
		case <-ctx.Done():
			if err := t.Stop(); err != nil {
				slog.Warn("Error stopping io-thread:", slog.String("id", t.id), slog.Any("error", err))
			}
			return ctx.Err()
		case err := <-errChan:
			return t.handleError(err)
		case cmdReq := <-t.adhocReqChan:
			t.handleCmdRequestWithTimeout(ctx, errChan, []*cmd.DiceDBCmd{cmdReq}, true, defaultRequestTimeout)
		case data := <-incomingDataChan:
			if err := t.processIncomingData(ctx, &data, errChan); err != nil {
				return err
			}
		case err := <-readErrChan:
			slog.Debug("Read error in io-thread, connection closed possibly", slog.String("id", t.id), slog.Any("error", err))
			return err
		}
	}
}

// startInputReader continuously reads input data from the ioHandler and sends it to the incomingDataChan.
func (t *BaseIOThread) startInputReader(ctx context.Context, incomingDataChan chan []byte, readErrChan chan error) {
	defer close(incomingDataChan)
	defer close(readErrChan)

	for {
		data, err := t.ioHandler.Read(ctx)
		if err != nil {
			select {
			case readErrChan <- err:
			case <-ctx.Done():
			}
			return
		}

		select {
		case incomingDataChan <- data:
		case <-ctx.Done():
			return
		}
	}
}

func (t *BaseIOThread) handleError(err error) error {
	if err != nil {
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
			slog.Debug("Connection closed for io-thread", slog.String("id", t.id), slog.Any("error", err))
			return err
		}
	}
	return fmt.Errorf("error writing response: %v", err)
}

func (t *BaseIOThread) processIncomingData(ctx context.Context, data *[]byte, errChan chan error) error {
	commands, err := t.parser.Parse(*data)

	if err != nil {
		err = t.ioHandler.Write(ctx, err)
		if err != nil {
			slog.Debug("Write error, connection closed possibly", slog.String("id", t.id), slog.Any("error", err))
			return err
		}
		return nil
	}

	if len(commands) == 0 {
		err = t.ioHandler.Write(ctx, fmt.Errorf("ERR: Invalid request"))
		if err != nil {
			slog.Debug("Write error, connection closed possibly", slog.String("id", t.id), slog.Any("error", err))
			return err
		}
		return nil
	}

	// DiceDB supports clients to send only one request at a time
	// We also need to ensure that the client is blocked until the response is received
	if len(commands) > 1 {
		err = t.ioHandler.Write(ctx, fmt.Errorf("ERR: Multiple commands not supported"))
		if err != nil {
			slog.Debug("Write error, connection closed possibly", slog.String("id", t.id), slog.Any("error", err))
			return err
		}
	}

	err = t.isAuthenticated(commands[0])
	if err != nil {
		writeErr := t.ioHandler.Write(ctx, err)
		if writeErr != nil {
			slog.Debug("Write error, connection closed possibly", slog.Any("error", errors.Join(err, writeErr)))
			return errors.Join(err, writeErr)
		}
		return nil
	}

	t.handleCmdRequestWithTimeout(ctx, errChan, commands, false, defaultRequestTimeout)
	return nil
}

func (t *BaseIOThread) handleCmdRequestWithTimeout(ctx context.Context, errChan chan error, commands []*cmd.DiceDBCmd, isWatchNotification bool, timeout time.Duration) {
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	t.executeCommandHandler(execCtx, errChan, commands, isWatchNotification)
}

func (t *BaseIOThread) executeCommandHandler(execCtx context.Context, errChan chan error, commands []*cmd.DiceDBCmd, isWatchNotification bool) {
	// Retrieve metadata for the command to determine if multisharding is supported.
	meta, ok := CommandsMeta[commands[0].Cmd]
	if ok && meta.preProcessing {
		if err := meta.preProcessResponse(t, commands[0]); err != nil {
			e := t.ioHandler.Write(execCtx, err)
			if e != nil {
				slog.Debug("Error executing for io-thread", slog.String("id", t.id), slog.Any("error", err))
			}
		}
	}

	err := t.executeCommand(execCtx, commands[0], isWatchNotification)
	if err != nil {
		slog.Error("Error executing command", slog.String("id", t.id), slog.Any("error", err))
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ETIMEDOUT) {
			slog.Debug("Connection closed for io-thread", slog.String("id", t.id), slog.Any("error", err))
			errChan <- err
		}
	}
}

func (t *BaseIOThread) executeCommand(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, isWatchNotification bool) error {
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
			err := t.ioHandler.Write(ctx, meta.IOThreadHandler(diceDBCmd.Args))
			slog.Debug("Error executing command on io-thread", slog.String("id", t.id), slog.Any("error", err))
			return err

		case SingleShard:
			// For single-shard or custom commands, process them without breaking up.
			cmdList = append(cmdList, diceDBCmd)

		case MultiShard, AllShard:
			var err error
			// If the command supports multisharding, break it down into multiple commands.
			cmdList, err = meta.decomposeCommand(ctx, t, diceDBCmd)
			if err != nil {
				var ioErr error
				// Check if it's a CustomError
				var customErr *diceerrors.PreProcessError
				if errors.As(err, &customErr) {
					ioErr = t.ioHandler.Write(ctx, customErr.Result)
				} else {
					ioErr = t.ioHandler.Write(ctx, err)
				}
				if ioErr != nil {
					slog.Debug("Error executing for io-thread", slog.String("id", t.id), slog.Any("error", ioErr))
				}
				return ioErr
			}

		case Custom:
			return t.handleCustomCommands(ctx, diceDBCmd)

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
		return t.handleCommandUnwatch(ctx, cmdList)
	}

	// Scatter the broken-down commands to the appropriate shards.
	if err := t.scatter(ctx, cmdList, meta.CmdType); err != nil {
		return err
	}

	// Gather the responses from the shards and write them to the buffer.
	if err := t.gather(ctx, diceDBCmd, len(cmdList), isWatchNotification, watchLabel); err != nil {
		return err
	}

	if meta.CmdType == Watch {
		// Proceed to subscribe after successful execution
		t.handleCommandWatch(cmdList)
	}

	return nil
}

func (t *BaseIOThread) handleCustomCommands(ctx context.Context, diceDBCmd *cmd.DiceDBCmd) error {
	// if command is of type Custom, write a custom logic around it
	switch diceDBCmd.Cmd {
	case CmdAuth:
		err := t.ioHandler.Write(ctx, t.RespAuth(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending auth response to io-thread", slog.String("id", t.id), slog.Any("error", err))
		}
		return err
	case CmdEcho:
		err := t.ioHandler.Write(ctx, RespEcho(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending echo response to io-thread", slog.String("id", t.id), slog.Any("error", err))
		}
		return err
	case CmdAbort:
		err := t.ioHandler.Write(ctx, clientio.OK)
		if err != nil {
			slog.Error("Error sending abort response to io-thread", slog.String("id", t.id), slog.Any("error", err))
		}
		slog.Info("Received ABORT command, initiating server shutdown", slog.String("id", t.id))
		t.globalErrorChan <- diceerrors.ErrAborted
		return err
	case CmdPing:
		err := t.ioHandler.Write(ctx, RespPING(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending ping response to io-thread", slog.String("id", t.id), slog.Any("error", err))
		}
		return err
	case CmdHello:
		err := t.ioHandler.Write(ctx, RespHello(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending ping response to io-thread", slog.String("id", t.id), slog.Any("error", err))
		}
		return err
	case CmdSleep:
		err := t.ioHandler.Write(ctx, RespSleep(diceDBCmd.Args))
		if err != nil {
			slog.Error("Error sending ping response to io-thread", slog.String("id", t.id), slog.Any("error", err))
		}
		return err
	default:
		return diceerrors.ErrUnknownCmd(diceDBCmd.Cmd)
	}
}

// handleCommandWatch sends a watch subscription request to the watch manager.
func (t *BaseIOThread) handleCommandWatch(cmdList []*cmd.DiceDBCmd) {
	t.cmdWatchSubscriptionChan <- watchmanager.WatchSubscription{
		Subscribe:    true,
		WatchCmd:     cmdList[len(cmdList)-1],
		AdhocReqChan: t.adhocReqChan,
	}
}

// handleCommandUnwatch sends an unwatch subscription request to the watch manager. It also sends a response to the client.
// The response is sent before the unwatch request is processed by the watch manager.
func (t *BaseIOThread) handleCommandUnwatch(ctx context.Context, cmdList []*cmd.DiceDBCmd) error {
	// extract the fingerprint
	command := cmdList[len(cmdList)-1]
	fp, parseErr := strconv.ParseUint(command.Args[0], 10, 32)
	if parseErr != nil {
		err := t.ioHandler.Write(ctx, diceerrors.ErrInvalidFingerprint)
		if err != nil {
			return fmt.Errorf("error sending push response to client: %v", err)
		}
		return parseErr
	}

	// send the unsubscribe request
	t.cmdWatchSubscriptionChan <- watchmanager.WatchSubscription{
		Subscribe:    false,
		AdhocReqChan: t.adhocReqChan,
		Fingerprint:  uint32(fp),
	}

	err := t.ioHandler.Write(ctx, clientio.RespOK)
	if err != nil {
		return fmt.Errorf("error sending push response to client: %v", err)
	}
	return nil
}

// scatter distributes the DiceDB commands to the respective shards based on the key.
// For each command, it calculates the shard ID and sends the command to the shard's request channel for processing.
func (t *BaseIOThread) scatter(ctx context.Context, cmds []*cmd.DiceDBCmd, cmdType CmdType) error {
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
			for i := uint8(0); i < uint8(t.shardManager.GetShardCount()); i++ {
				// Get the shard ID (i) and its associated request channel.
				shardID, responseChan := i, t.shardManager.GetShard(i).ReqChan

				// Send a StoreOp operation to the shard's request channel.
				responseChan <- &ops.StoreOp{
					SeqID:      i,                         // Sequence ID for this operation.
					RequestID:  GenerateUniqueRequestID(), // Unique identifier for the request.
					Cmd:        cmds[0],                   // Command to be executed, using the first command in cmds.
					IOThreadID: t.id,                      // ID of the current io-thread.
					ShardID:    shardID,                   // ID of the shard handling this operation.
					Client:     nil,                       // Client information (if applicable).
				}
			}
		} else {
			// If the command type is specific to certain commands, process them individually.
			for i := uint8(0); i < uint8(len(cmds)); i++ {
				// Determine the appropriate shard for the current command using a routing key.
				shardID, responseChan := t.shardManager.GetShardInfo(getRoutingKeyFromCommand(cmds[i]))

				// Send a StoreOp operation to the shard's request channel.
				responseChan <- &ops.StoreOp{
					SeqID:      i,                         // Sequence ID for this operation.
					RequestID:  GenerateUniqueRequestID(), // Unique identifier for the request.
					Cmd:        cmds[i],                   // Command to be executed, using the current command in cmds.
					IOThreadID: t.id,                      // ID of the current io-thread.
					ShardID:    shardID,                   // ID of the shard handling this operation.
					Client:     nil,                       // Client information (if applicable).
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
func (t *BaseIOThread) gather(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, numCmds int, isWatchNotification bool, watchLabel string) error {
	// Collect responses from all shards
	storeOp, err := t.gatherResponses(ctx, numCmds)
	if err != nil {
		return err
	}

	if len(storeOp) == 0 {
		slog.Error("No response from shards",
			slog.String("id", t.id),
			slog.String("command", diceDBCmd.Cmd))
		return fmt.Errorf("no response from shards for command: %s", diceDBCmd.Cmd)
	}

	if isWatchNotification {
		return t.handleWatchNotification(ctx, diceDBCmd, storeOp[0], watchLabel)
	}

	// Process command based on its type
	cmdMeta, ok := CommandsMeta[diceDBCmd.Cmd]
	if !ok {
		return t.handleUnsupportedCommand(ctx, storeOp[0])
	}

	return t.handleCommand(ctx, cmdMeta, diceDBCmd, storeOp)
}

// gatherResponses collects responses from all shards
func (t *BaseIOThread) gatherResponses(ctx context.Context, numCmds int) ([]ops.StoreResponse, error) {
	storeOp := make([]ops.StoreResponse, 0, numCmds)

	for numCmds > 0 {
		select {
		case <-ctx.Done():
			slog.Error("Timed out waiting for response from shards",
				slog.String("id", t.id),
				slog.Any("error", ctx.Err()))
			return nil, ctx.Err()

		case resp, ok := <-t.responseChan:
			if ok {
				storeOp = append(storeOp, *resp)
			}
			numCmds--

		case sError, ok := <-t.shardManager.ShardErrorChan:
			if ok {
				slog.Error("Error from shard",
					slog.String("id", t.id),
					slog.Any("error", sError))
				return nil, sError.Error
			}
		}
	}

	return storeOp, nil
}

// handleWatchNotification processes watch notification responses
func (t *BaseIOThread) handleWatchNotification(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, resp ops.StoreResponse, watchLabel string) error {
	fingerprint := fmt.Sprintf("%d", diceDBCmd.GetFingerprint())

	// if watch label is not empty, then this is the first response for the watch command
	// hence, we will send the watch label as part of the response
	firstRespElem := diceDBCmd.Cmd
	if watchLabel != "" {
		firstRespElem = watchLabel
	}

	if resp.EvalResponse.Error != nil {
		return t.writeResponse(ctx, querymanager.GenericWatchResponse(firstRespElem, fingerprint, resp.EvalResponse.Error))
	}

	return t.writeResponse(ctx, querymanager.GenericWatchResponse(firstRespElem, fingerprint, resp.EvalResponse.Result))
}

// handleUnsupportedCommand processes commands not in CommandsMeta
func (t *BaseIOThread) handleUnsupportedCommand(ctx context.Context, resp ops.StoreResponse) error {
	if resp.EvalResponse.Error != nil {
		return t.writeResponse(ctx, resp.EvalResponse.Error)
	}
	return t.writeResponse(ctx, resp.EvalResponse.Result)
}

// handleCommand processes commands based on their type
func (t *BaseIOThread) handleCommand(ctx context.Context, cmdMeta CmdMeta, diceDBCmd *cmd.DiceDBCmd, storeOp []ops.StoreResponse) error {
	var err error

	switch cmdMeta.CmdType {
	case SingleShard, Custom:
		if storeOp[0].EvalResponse.Error != nil {
			err = t.writeResponse(ctx, storeOp[0].EvalResponse.Error)
		} else {
			err = t.writeResponse(ctx, storeOp[0].EvalResponse.Result)
		}

		if err == nil && t.wl != nil {
			t.wl.LogCommand(diceDBCmd)
		}
	case MultiShard, AllShard:
		err = t.writeResponse(ctx, cmdMeta.composeResponse(storeOp...))

		if err == nil && t.wl != nil {
			t.wl.LogCommand(diceDBCmd)
		}
	default:
		slog.Error("Unknown command type",
			slog.String("id", t.id),
			slog.String("command", diceDBCmd.Cmd),
			slog.Any("evalResp", storeOp))
		err = t.writeResponse(ctx, diceerrors.ErrInternalServer)
	}
	return err
}

// writeResponse handles writing responses and logging errors
func (t *BaseIOThread) writeResponse(ctx context.Context, response interface{}) error {
	err := t.ioHandler.Write(ctx, response)
	if err != nil {
		slog.Debug("Error sending response to client",
			slog.String("id", t.id),
			slog.Any("error", err))
	}
	return err
}

func (t *BaseIOThread) isAuthenticated(diceDBCmd *cmd.DiceDBCmd) error {
	if diceDBCmd.Cmd != auth.Cmd && !t.Session.IsActive() {
		return errors.New("NOAUTH Authentication required")
	}

	return nil
}

func (t *BaseIOThread) Stop() error {
	slog.Info("Stopping io-thread", slog.String("id", t.id))
	t.Session.Expire()
	return nil
}

func GenerateUniqueRequestID() uint32 {
	return atomic.AddUint32(&requestCounter, 1)
}
