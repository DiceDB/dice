package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/server/abstractserver"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio"
	"github.com/dicedb/dice/internal/clientio/iohandler/netconn"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/iomultiplexer"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/querymanager"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
)

type AsyncServer struct {
	abstractserver.AbstractServer
	serverFD               int
	maxClients             int32
	multiplexer            iomultiplexer.IOMultiplexer
	multiplexerPollTimeout time.Duration
	connectedClients       map[int]*comm.Client
	queryWatcher           *querymanager.Manager
	shardManager           *shard.ShardManager
	ioChan                 chan *ops.StoreResponse     // The server acts like a worker today, this behavior will change once IOThreads are introduced and each client gets its own worker.
	queryWatchChan         chan dstore.QueryWatchEvent // This is needed to co-ordinate between the store and the query watcher.
}

// NewAsyncServer initializes a new AsyncServer
func NewAsyncServer(shardManager *shard.ShardManager, queryWatchChan chan dstore.QueryWatchEvent) *AsyncServer {
	return &AsyncServer{
		maxClients:             config.DiceConfig.Performance.MaxClients,
		connectedClients:       make(map[int]*comm.Client),
		shardManager:           shardManager,
		queryWatcher:           querymanager.NewQueryManager(),
		multiplexerPollTimeout: config.DiceConfig.Performance.MultiplexerPollTimeout,
		ioChan:                 make(chan *ops.StoreResponse, 1000),
		queryWatchChan:         queryWatchChan,
	}
}

// SetupUsers initializes the default user for the server
func (s *AsyncServer) SetupUsers() error {
	user, err := auth.UserStore.Add(config.DiceConfig.Auth.UserName)
	if err != nil {
		return err
	}
	return user.SetPassword(config.DiceConfig.Auth.Password)
}

// FindPortAndBind binds the server to the given host and port
func (s *AsyncServer) FindPortAndBind() (socketErr error) {
	serverFD, socketErr := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)

	if socketErr != nil {
		return socketErr
	}

	// Close the socket on exit if an error occurs
	defer func() {
		if socketErr != nil {
			if err := syscall.Close(serverFD); err != nil {
				slog.Warn("failed to close server socket", slog.Any("error", err))
			}
		}
	}()

	if err := syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return err
	}

	s.serverFD = serverFD

	if err := syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	ip4 := net.ParseIP(config.DiceConfig.AsyncServer.Addr)
	if ip4 == nil {
		return diceerrors.ErrInvalidIPAddress
	}

	return syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.DiceConfig.AsyncServer.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	})
}

// ClosePort ensures the server socket is closed properly.
func (s *AsyncServer) ClosePort() {
	if s.serverFD != 0 {
		if err := syscall.Close(s.serverFD); err != nil {
			slog.Warn("failed to close server socket", slog.Any("error", err))
		}
		s.serverFD = 0
	}
}

// InitiateShutdown gracefully shuts down the server
func (s *AsyncServer) InitiateShutdown() {
	// Close the server socket first
	s.ClosePort()

	s.shardManager.UnregisterWorker("server")

	// Close all client connections
	for fd := range s.connectedClients {
		if err := syscall.Close(fd); err != nil {
			slog.Warn("failed to close client connection", slog.Any("error", err))
		}
		delete(s.connectedClients, fd)
	}
}

// Run starts the server, accepts connections, and handles client requests
func (s *AsyncServer) Run(ctx context.Context) error {
	if err := s.SetupUsers(); err != nil {
		return err
	}

	watchCtx, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.queryWatcher.Run(watchCtx, s.queryWatchChan)
	}()

	s.shardManager.RegisterWorker("server", s.ioChan, nil)

	if err := syscall.Listen(s.serverFD, int(s.maxClients)); err != nil {
		return err
	}

	var err error
	s.multiplexer, err = iomultiplexer.New(s.maxClients)
	if err != nil {
		return err
	}

	defer func() {
		if err := s.multiplexer.Close(); err != nil {
			slog.Warn("failed to close multiplexer", slog.Any("error", err))
		}
	}()

	if err := s.multiplexer.Subscribe(iomultiplexer.Event{
		Fd: s.serverFD,
		Op: iomultiplexer.OpRead,
	}); err != nil {
		return err
	}

	eventLoopCtx, cancelEventLoop := context.WithCancel(ctx)
	defer cancelEventLoop()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = s.eventLoop(eventLoopCtx)
		if err != nil {
			cancelWatch()
			cancelEventLoop()
			s.InitiateShutdown()
		}
	}()

	wg.Wait()

	return err
}

// eventLoop listens for events and handles client requests. It also runs a cron job to delete expired keys
func (s *AsyncServer) eventLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			events, err := s.multiplexer.Poll(s.multiplexerPollTimeout)
			if err != nil {
				if errors.Is(err, syscall.EINTR) {
					continue
				}
				return err
			}

			for _, event := range events {
				if event.Fd == s.serverFD {
					if err := s.acceptConnection(); err != nil {
						slog.Warn(err.Error())
					}
				} else {
					if err := s.handleClientEvent(event); err != nil {
						if errors.Is(err, diceerrors.ErrAborted) {
							slog.Debug("Received abort command, initiating graceful shutdown")
							return err
						} else if !errors.Is(err, syscall.ECONNRESET) && !errors.Is(err, net.ErrClosed) {
							slog.Warn(err.Error())
						}
					}
				}
			}
		}
	}
}

// acceptConnection accepts a new client connection and subscribes to read events on the connection.
func (s *AsyncServer) acceptConnection() error {
	fd, _, err := syscall.Accept(s.serverFD)
	if err != nil {
		return err
	}

	s.connectedClients[fd] = comm.NewClient(fd)
	if err := syscall.SetNonblock(fd, true); err != nil {
		return err
	}

	return s.multiplexer.Subscribe(iomultiplexer.Event{
		Fd: fd,
		Op: iomultiplexer.OpRead,
	})
}

// handleClientEvent reads commands from the client connection and responds to the client. It also handles disconnections.
func (s *AsyncServer) handleClientEvent(event iomultiplexer.Event) error {
	client := s.connectedClients[event.Fd]
	if client == nil {
		return nil
	}

	commands, hasAbort, err := readCommands(client)
	if err != nil {
		if err := syscall.Close(event.Fd); err != nil {
			slog.Error("error closing client connection", slog.Any("error", err))
		}
		delete(s.connectedClients, event.Fd)
		return err
	}

	s.EvalAndRespond(commands, client)
	if hasAbort {
		return diceerrors.ErrAborted
	}

	return nil
}

func handleMigratedResp(resp interface{}, buf *bytes.Buffer) {
	// Process the incoming response by calling the handleResponse function.
	// This function checks the response against known RESP formatted values
	// and returns the corresponding byte array representation. The result
	// is assigned to the resp variable.
	r := netconn.HandlePredefinedResponse(resp)

	// Check if the processed response (resp) is not nil.
	// If it is not nil, this means incoming response was not
	// matched to any predefined RESP responses,
	// and we proceed to encode the original response using
	// the clientio.Encode function. This function converts the
	// response into the desired format based on the specified
	// isBlkEnc encoding flag, which indicates whether the
	// response should be encoded in a block format.
	if r == nil {
		r = clientio.Encode(resp, false)
	}

	buf.Write(r)
}

func (s *AsyncServer) executeCommandToBuffer(diceDBCmd *cmd.DiceDBCmd, buf *bytes.Buffer, c *comm.Client) {
	s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
		Cmd:      diceDBCmd,
		WorkerID: "server",
		ShardID:  0,
		Client:   c,
	}

	resp := <-s.ioChan

	val, ok := WorkerCmdsMeta[diceDBCmd.Cmd]
	// TODO: Remove this conditional check and if (true) condition when all commands are migrated
	if !ok {
		buf.Write(resp.EvalResponse.Result.([]byte))
	} else {
		// If command type is Global then return the worker eval
		if val.CmdType == Global {
			buf.Write(val.RespNoShards(diceDBCmd.Args))
			return
		}
		// Handle error case independently
		if resp.EvalResponse.Error != nil {
			handleMigratedResp(resp.EvalResponse.Error, buf)
			return
		}
		handleMigratedResp(resp.EvalResponse.Result, buf)
		return
	}
}

func readCommands(c io.ReadWriter) (*cmd.RedisCmds, bool, error) {
	var hasABORT = false
	rp := clientio.NewRESPParser(c)
	values, err := rp.DecodeMultiple()
	if err != nil {
		return nil, false, err
	}

	var cmds = make([]*cmd.DiceDBCmd, 0)
	for _, value := range values {
		arrayValue, ok := value.([]interface{})
		if !ok {
			return nil, false, fmt.Errorf("expected array, got %T", value)
		}

		tokens, err := toArrayString(arrayValue)
		if err != nil {
			return nil, false, err
		}

		if len(tokens) == 0 {
			return nil, false, fmt.Errorf("empty command")
		}

		command := strings.ToUpper(tokens[0])
		cmds = append(cmds, &cmd.DiceDBCmd{
			Cmd:  command,
			Args: tokens[1:],
		})

		if command == "ABORT" {
			hasABORT = true
		}
	}

	rCmds := &cmd.RedisCmds{
		Cmds: cmds,
	}
	return rCmds, hasABORT, nil
}

func toArrayString(ai []interface{}) ([]string, error) {
	as := make([]string, len(ai))
	for i, v := range ai {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string", i)
		}
		as[i] = s
	}
	return as, nil
}

func (s *AsyncServer) EvalAndRespond(cmds *cmd.RedisCmds, c *comm.Client) {
	var resp []byte
	buf := bytes.NewBuffer(resp)

	for _, diceDBCmd := range cmds.Cmds {
		if !s.isAuthenticated(diceDBCmd, c, buf) {
			continue
		}

		if c.IsTxn {
			s.handleTransactionCommand(diceDBCmd, c, buf)
		} else {
			s.handleNonTransactionCommand(diceDBCmd, c, buf)
		}
	}

	s.writeResponse(c, buf)
}

func (s *AsyncServer) isAuthenticated(diceDBCmd *cmd.DiceDBCmd, c *comm.Client, buf *bytes.Buffer) bool {
	if diceDBCmd.Cmd != auth.Cmd && !c.Session.IsActive() {
		buf.Write(clientio.Encode(errors.New("NOAUTH Authentication required"), false))
		return false
	}
	return true
}

func (s *AsyncServer) handleTransactionCommand(diceDBCmd *cmd.DiceDBCmd, c *comm.Client, buf *bytes.Buffer) {
	if eval.TxnCommands[diceDBCmd.Cmd] {
		switch diceDBCmd.Cmd {
		case eval.ExecCmdMeta.Name:
			s.executeTransaction(c, buf)
		case eval.DiscardCmdMeta.Name:
			s.discardTransaction(c, buf)
		default:
			slog.Error(
				"Unhandled transaction command",
				slog.String("command", diceDBCmd.Cmd),
			)
		}
	} else {
		c.TxnQueue(diceDBCmd)
		buf.Write(clientio.RespQueued)
	}
}

func (s *AsyncServer) handleNonTransactionCommand(diceDBCmd *cmd.DiceDBCmd, c *comm.Client, buf *bytes.Buffer) {
	switch diceDBCmd.Cmd {
	case eval.MultiCmdMeta.Name:
		c.TxnBegin()
		buf.Write(clientio.RespOK)
	case eval.ExecCmdMeta.Name:
		buf.Write(diceerrors.NewErrWithMessage("EXEC without MULTI"))
	case eval.DiscardCmdMeta.Name:
		buf.Write(diceerrors.NewErrWithMessage("DISCARD without MULTI"))
	default:
		s.executeCommandToBuffer(diceDBCmd, buf, c)
	}
}

func (s *AsyncServer) executeTransaction(c *comm.Client, buf *bytes.Buffer) {
	cmds := c.Cqueue.Cmds
	_, err := fmt.Fprintf(buf, "*%d\r\n", len(cmds))
	if err != nil {
		slog.Error("Error writing to buffer", slog.Any("error", err))
		return
	}

	for _, cmd := range cmds {
		s.executeCommandToBuffer(cmd, buf, c)
	}

	c.Cqueue.Cmds = make([]*cmd.DiceDBCmd, 0)
	c.IsTxn = false
}

func (s *AsyncServer) discardTransaction(c *comm.Client, buf *bytes.Buffer) {
	c.TxnDiscard()
	buf.Write(clientio.RespOK)
}

func (s *AsyncServer) writeResponse(c *comm.Client, buf *bytes.Buffer) {
	if _, err := c.Write(buf.Bytes()); err != nil {
		slog.Error(err.Error())
	}
}
