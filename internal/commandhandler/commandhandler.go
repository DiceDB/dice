// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package commandhandler

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

type CommandHandler interface {
	ID() string
	Start(ctx context.Context) error
	Stop() error
}

type BaseCommandHandler struct {
	CommandHandler
	id     string
	parser requestparser.Parser
	wl     wal.AbstractWAL

	shardManager             *shard.ShardManager
	Session                  *auth.Session
	adhocReqChan             chan *cmd.DiceDBCmd
	globalErrorChan          chan error
	ioThreadReadChan         chan []byte             // Channel to receive data from io-thread
	ioThreadWriteChan        chan interface{}        // Channel to send data to io-thread
	ioThreadErrChan          chan error              // Channel to receive errors from io-thread
	responseChan             chan *ops.StoreResponse // Channel to communicate with shard
	preprocessingChan        chan *ops.StoreResponse // Channel to communicate with shard
	cmdWatchSubscriptionChan chan watchmanager.WatchSubscription
}

func NewCommandHandler(id string, responseChan, preprocessingChan chan *ops.StoreResponse,
	cmdWatchSubscriptionChan chan watchmanager.WatchSubscription,
	parser requestparser.Parser, shardManager *shard.ShardManager, gec chan error,
	ioThreadReadChan chan []byte, ioThreadWriteChan chan interface{}, ioThreadErrChan chan error,
	wl wal.AbstractWAL) *BaseCommandHandler {
	return &BaseCommandHandler{
		id:                       id,
		parser:                   parser,
		shardManager:             shardManager,
		adhocReqChan:             make(chan *cmd.DiceDBCmd, config.DiceConfig.Performance.AdhocReqChanBufSize),
		Session:                  auth.NewSession(),
		globalErrorChan:          gec,
		ioThreadReadChan:         ioThreadReadChan,
		ioThreadWriteChan:        ioThreadWriteChan,
		ioThreadErrChan:          ioThreadErrChan,
		responseChan:             responseChan,
		preprocessingChan:        preprocessingChan,
		cmdWatchSubscriptionChan: cmdWatchSubscriptionChan,
		wl:                       wl,
	}
}

func (h *BaseCommandHandler) ID() string {
	return h.id
}

func (h *BaseCommandHandler) Start(ctx context.Context) error {
	errChan := make(chan error, 1) // for adhoc request processing errors

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-h.ioThreadErrChan:
			return err
		case cmdReq := <-h.adhocReqChan:
			resp, err := h.handleCmdRequestWithTimeout(ctx, errChan, []*cmd.DiceDBCmd{cmdReq}, true, defaultRequestTimeout)
			h.sendResponseToIOThread(resp, err)
		case err := <-errChan:
			return h.handleError(err)
		case data := <-h.ioThreadReadChan:
			resp, err := h.processCommand(ctx, &data, h.globalErrorChan)
			h.sendResponseToIOThread(resp, err)
		}
	}
}

// processCommand processes commands recevied from io thread
func (h *BaseCommandHandler) processCommand(ctx context.Context, data *[]byte, gec chan error) (interface{}, error) {
	commands, err := h.parser.Parse(*data)

	if err != nil {
		slog.Debug("error parsing commands from io thread", slog.String("id", h.id), slog.Any("error", err))
		return nil, err
	}

	if len(commands) == 0 {
		slog.Debug("invalid request from io thread with zero length", slog.String("id", h.id))
		return nil, fmt.Errorf("ERR: Invalid request")
	}

	// DiceDB supports clients to send only one request at a time
	// We also need to ensure that the client is blocked until the response is received
	if len(commands) > 1 {
		return nil, fmt.Errorf("ERR: Multiple commands not supported")
	}

	err = h.isAuthenticated(commands[0])
	if err != nil {
		slog.Debug("command handler authentication failed", slog.String("id", h.id), slog.Any("error", err))
		return nil, err
	}

	return h.handleCmdRequestWithTimeout(ctx, gec, commands, false, defaultRequestTimeout)
}

func (h *BaseCommandHandler) handleCmdRequestWithTimeout(ctx context.Context, gec chan error, commands []*cmd.DiceDBCmd, isWatchNotification bool, timeout time.Duration) (interface{}, error) {
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return h.executeCommandHandler(execCtx, gec, commands, isWatchNotification)
}

func (h *BaseCommandHandler) executeCommandHandler(execCtx context.Context, gec chan error, commands []*cmd.DiceDBCmd, isWatchNotification bool) (interface{}, error) {
	// Retrieve metadata for the command to determine if multisharding is supported.
	meta, ok := CommandsMeta[commands[0].Cmd]
	if ok && meta.preProcessing {
		if err := meta.preProcessResponse(h, commands[0]); err != nil {
			slog.Debug("error pre processing response", slog.String("id", h.id), slog.Any("error", err))
			return nil, err
		}
	}

	resp, err := h.executeCommand(execCtx, commands[0], isWatchNotification)

	// log error and send to global error channel if it's a connection error
	if err != nil {
		slog.Error("Error executing command", slog.String("id", h.id), slog.Any("error", err))
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ETIMEDOUT) {
			slog.Debug("Connection closed for io-thread", slog.String("id", h.id), slog.Any("error", err))
			gec <- err
		}
	}

	return resp, err
}

func (h *BaseCommandHandler) executeCommand(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, isWatchNotification bool) (interface{}, error) {
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
			// process global command immediately without involving any shards.
			return meta.CmdHandlerFunction(diceDBCmd.Args), nil

		case SingleShard:
			// For single-shard or custom commands, process them without breaking up.
			cmdList = append(cmdList, diceDBCmd)

		case MultiShard, AllShard:
			var err error
			// If the command supports multisharding, break it down into multiple commands.
			cmdList, err = meta.decomposeCommand(h, ctx, diceDBCmd)
			if err != nil {
				slog.Debug("error decomposing command", slog.String("id", h.id), slog.Any("error", err))
				// Check if it's a CustomError
				var customErr *diceerrors.PreProcessError
				if errors.As(err, &customErr) {
					return nil, err
				}
				return nil, err
			}

		case Custom:
			return h.handleCustomCommands(diceDBCmd)

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
		return h.handleCommandUnwatch(cmdList)
	}

	// Scatter the broken-down commands to the appropriate shards.
	if err := h.scatter(ctx, cmdList, meta.CmdType); err != nil {
		return nil, err
	}

	// Gather the responses from the shards and write them to the buffer.
	resp, err := h.gather(ctx, diceDBCmd, len(cmdList), isWatchNotification, watchLabel)
	if err != nil {
		return nil, err
	}

	// Proceed to subscribe after successful execution
	if meta.CmdType == Watch {
		h.handleCommandWatch(cmdList)
	}

	return resp, nil
}

func (h *BaseCommandHandler) handleCustomCommands(diceDBCmd *cmd.DiceDBCmd) (interface{}, error) {
	// if command is of type Custom, write a custom logic around it
	switch diceDBCmd.Cmd {
	case CmdAuth:
		return h.RespAuth(diceDBCmd.Args), nil
	case CmdEcho:
		return RespEcho(diceDBCmd.Args), nil
	case CmdAbort:
		slog.Info("Received ABORT command, initiating server shutdown", slog.String("id", h.id))
		h.globalErrorChan <- diceerrors.ErrAborted
		return clientio.OK, nil
	case CmdPing:
		return RespPING(diceDBCmd.Args), nil
	case CmdHello:
		return RespHello(diceDBCmd.Args), nil
	case CmdSleep:
		return RespSleep(diceDBCmd.Args), nil
	default:
		return nil, diceerrors.ErrUnknownCmd(diceDBCmd.Cmd)
	}
}

// handleCommandWatch sends a watch subscription request to the watch manager.
func (h *BaseCommandHandler) handleCommandWatch(cmdList []*cmd.DiceDBCmd) {
	h.cmdWatchSubscriptionChan <- watchmanager.WatchSubscription{
		Subscribe:    true,
		WatchCmd:     cmdList[len(cmdList)-1],
		AdhocReqChan: h.adhocReqChan,
	}
}

// handleCommandUnwatch sends an unwatch subscription request to the watch manager. It also sends a response to the client.
// The response is sent before the unwatch request is processed by the watch manager.
func (h *BaseCommandHandler) handleCommandUnwatch(cmdList []*cmd.DiceDBCmd) (interface{}, error) {
	// extract the fingerprint
	command := cmdList[len(cmdList)-1]
	fp, parseErr := strconv.ParseUint(command.Args[0], 10, 32)
	if parseErr != nil {
		return nil, diceerrors.ErrInvalidFingerprint
	}

	// send the unsubscribe request
	h.cmdWatchSubscriptionChan <- watchmanager.WatchSubscription{
		Subscribe:    false,
		AdhocReqChan: h.adhocReqChan,
		Fingerprint:  uint32(fp),
	}

	return clientio.OK, nil
}

// scatter distributes the DiceDB commands to the respective shards based on the key.
// For each command, it calculates the shard ID and sends the command to the shard's request channel for processing.
func (h *BaseCommandHandler) scatter(ctx context.Context, cmds []*cmd.DiceDBCmd, cmdType CmdType) error {
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
			for i := uint8(0); i < uint8(h.shardManager.GetShardCount()); i++ {
				// Get the shard ID (i) and its associated request channel.
				shardID, responseChan := i, h.shardManager.GetShard(i).ReqChan

				// Send a StoreOp operation to the shard's request channel.
				responseChan <- &ops.StoreOp{
					SeqID:        i,                         // Sequence ID for this operation.
					RequestID:    GenerateUniqueRequestID(), // Unique identifier for the request.
					Cmd:          cmds[0],                   // Command to be executed, using the first command in cmds.
					CmdHandlerID: h.id,                      // ID of the current command handler.
					ShardID:      shardID,                   // ID of the shard handling this operation.
					Client:       nil,                       // Client information (if applicable).
				}
			}
		} else {
			// If the command type is specific to certain commands, process them individually.
			for i := uint8(0); i < uint8(len(cmds)); i++ {
				// Determine the appropriate shard for the current command using a routing key.
				shardID, responseChan := h.shardManager.GetShardInfo(getRoutingKeyFromCommand(cmds[i]))

				// Send a StoreOp operation to the shard's request channel.
				responseChan <- &ops.StoreOp{
					SeqID:        i,                         // Sequence ID for this operation.
					RequestID:    GenerateUniqueRequestID(), // Unique identifier for the request.
					Cmd:          cmds[i],                   // Command to be executed, using the current command in cmds.
					CmdHandlerID: h.id,                      // ID of the current command handler.
					ShardID:      shardID,                   // ID of the shard handling this operation.
					Client:       nil,                       // Client information (if applicable).
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
func (h *BaseCommandHandler) gather(ctx context.Context, diceDBCmd *cmd.DiceDBCmd, numCmds int, isWatchNotification bool, watchLabel string) (interface{}, error) {
	// Collect responses from all shards
	storeOp, err := h.gatherResponses(ctx, numCmds)
	if err != nil {
		return nil, err
	}

	if len(storeOp) == 0 {
		slog.Error("No response from shards",
			slog.String("id", h.id),
			slog.String("command", diceDBCmd.Cmd))
		return nil, fmt.Errorf("no response from shards for command: %s", diceDBCmd.Cmd)
	}

	if isWatchNotification {
		return h.handleWatchNotification(diceDBCmd, storeOp[0], watchLabel)
	}

	// Process command based on its type
	cmdMeta, ok := CommandsMeta[diceDBCmd.Cmd]
	if !ok {
		return h.handleUnsupportedCommand(storeOp[0])
	}

	return h.handleCommand(cmdMeta, diceDBCmd, storeOp)
}

// gatherResponses collects responses from all shards
func (h *BaseCommandHandler) gatherResponses(ctx context.Context, numCmds int) ([]ops.StoreResponse, error) {
	storeOp := make([]ops.StoreResponse, 0, numCmds)

	for numCmds > 0 {
		select {
		case <-ctx.Done():
			slog.Error("Timed out waiting for response from shards",
				slog.String("id", h.id),
				slog.Any("error", ctx.Err()))
			return nil, ctx.Err()

		case resp, ok := <-h.responseChan:
			if ok {
				storeOp = append(storeOp, *resp)
			}
			numCmds--

		case sError, ok := <-h.shardManager.ShardErrorChan:
			if ok {
				slog.Error("Error from shard",
					slog.String("id", h.id),
					slog.Any("error", sError))
				return nil, sError.Error
			}
		}
	}

	return storeOp, nil
}

// handleWatchNotification processes watch notification responses
func (h *BaseCommandHandler) handleWatchNotification(diceDBCmd *cmd.DiceDBCmd, resp ops.StoreResponse, watchLabel string) (interface{}, error) {
	fingerprint := fmt.Sprintf("%d", diceDBCmd.GetFingerprint())

	// if watch label is not empty, then this is the first response for the watch command
	// hence, we will send the watch label as part of the response
	firstRespElem := diceDBCmd.Cmd
	if watchLabel != "" {
		firstRespElem = watchLabel
	}

	if resp.EvalResponse.Error != nil {
		// This is a special case where error is returned as part of the watch response
		return querymanager.GenericWatchResponse(firstRespElem, fingerprint, resp.EvalResponse.Error), nil
	}

	return querymanager.GenericWatchResponse(firstRespElem, fingerprint, resp.EvalResponse.Result), nil
}

// handleUnsupportedCommand processes commands not in CommandsMeta
func (h *BaseCommandHandler) handleUnsupportedCommand(resp ops.StoreResponse) (interface{}, error) {
	if resp.EvalResponse.Error != nil {
		return nil, resp.EvalResponse.Error
	}
	return resp.EvalResponse.Result, nil
}

// handleCommand processes commands based on their type
func (h *BaseCommandHandler) handleCommand(cmdMeta CmdMeta, diceDBCmd *cmd.DiceDBCmd, storeOp []ops.StoreResponse) (interface{}, error) {
	switch cmdMeta.CmdType {
	case SingleShard, Custom:
		if storeOp[0].EvalResponse.Error != nil {
			return nil, storeOp[0].EvalResponse.Error
		} else {
			return storeOp[0].EvalResponse.Result, nil
		}

	case MultiShard, AllShard:
		return cmdMeta.composeResponse(storeOp...), nil

	default:
		slog.Error("Unknown command type",
			slog.String("id", h.id),
			slog.String("command", diceDBCmd.Cmd),
			slog.Any("evalResp", storeOp))
		return nil, diceerrors.ErrInternalServer
	}
}

func (h *BaseCommandHandler) handleError(err error) error {
	if err != nil {
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
			slog.Debug("Connection closed for io-thread", slog.String("id", h.id), slog.Any("error", err))
			return err
		}
	}
	return fmt.Errorf("error writing response: %v", err)
}

func (h *BaseCommandHandler) sendResponseToIOThread(resp interface{}, err error) {
	if err != nil {
		var customErr *diceerrors.PreProcessError
		if errors.As(err, &customErr) {
			h.ioThreadWriteChan <- customErr.Result
		}
		h.ioThreadWriteChan <- err
		return
	}
	h.ioThreadWriteChan <- resp
}

func (h *BaseCommandHandler) isAuthenticated(diceDBCmd *cmd.DiceDBCmd) error {
	if diceDBCmd.Cmd != auth.Cmd && !h.Session.IsActive() {
		return errors.New("NOAUTH Authentication required")
	}

	return nil
}

func (h *BaseCommandHandler) Stop() error {
	slog.Info("Stopping command handler", slog.String("id", h.id))
	h.Session.Expire()
	return nil
}

func GenerateUniqueRequestID() uint32 {
	return atomic.AddUint32(&requestCounter, 1)
}

// RespAuth returns with an encoded "OK" if the user is authenticated
// If the user is not authenticated, it returns with an encoded error message
func (h *BaseCommandHandler) RespAuth(args []string) interface{} {
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

	if err := h.Session.Validate(username, password); err != nil {
		return err
	}

	return clientio.OK
}
