package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/dicedb/dice/core/auth"
	"github.com/dicedb/dice/core/cmd"
	"github.com/dicedb/dice/core/comm"
	"github.com/dicedb/dice/core/ops"

	"github.com/dicedb/dice/core/diceerrors"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/core/iomultiplexer"
	"github.com/dicedb/dice/internal/constants"
)

var ErrAborted = errors.New("server received ABORT command")
var ErrInvalidIPAddress = errors.New("invalid IP address")

type AsyncServer struct {
	serverFD               int
	maxClients             int
	multiplexer            iomultiplexer.IOMultiplexer
	multiplexerPollTimeout time.Duration
	connectedClients       map[int]*comm.Client
	queryWatcher           *core.QueryWatcher
	shardManager           *core.ShardManager
	ioChan                 chan *ops.StoreResponse // The server acts like a worker today, this behavior will change once IOThreads are introduced and each client gets its own worker.
}

// NewAsyncServer initializes a new AsyncServer
func NewAsyncServer() *AsyncServer {
	shardManager := core.NewShardManager(1)
	return &AsyncServer{
		maxClients:             config.ServerMaxClients,
		connectedClients:       make(map[int]*comm.Client),
		shardManager:           shardManager,
		queryWatcher:           core.NewQueryWatcher(shardManager),
		multiplexerPollTimeout: config.ServerMultiplexerPollTimeout,
		ioChan:                 make(chan *ops.StoreResponse, 1000),
	}
}

// SetupUsers initializes the default user for the server
func (s *AsyncServer) SetupUsers() error {
	user, err := auth.UserStore.Add(auth.DefaultUserName)
	if err != nil {
		return err
	}
	if err := user.SetPassword(config.RequirePass); err != nil {
		return err
	}
	log.Info("default user set up", "password required", config.RequirePass != constants.EmptyStr)
	return nil
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
				log.Warn("failed to close server socket", "error", err)
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

	ip4 := net.ParseIP(config.Host)
	if ip4 == nil {
		return ErrInvalidIPAddress
	}

	return syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	})
}

// ClosePort ensures the server socket is closed properly.
func (s *AsyncServer) ClosePort() {
	if s.serverFD != 0 {
		if err := syscall.Close(s.serverFD); err != nil {
			log.Warn("failed to close server socket", "error", err)
		} else {
			log.Info("Server socket closed successfully")
		}
		s.serverFD = 0
	}
}

// WaitForSignal listens for OS signals and triggers shutdown
func (s *AsyncServer) WaitForSignal(cancel context.CancelFunc, sigs chan os.Signal) {
	sig := <-sigs
	log.Info("Signal received, initiating shutdown", "signal", sig)
	cancel()
	s.InitiateShutdown()
}

// InitiateShutdown gracefully shuts down the server
func (s *AsyncServer) InitiateShutdown() {
	// Close the server socket first
	s.ClosePort()

	// Close all client connections
	for fd := range s.connectedClients {
		if err := syscall.Close(fd); err != nil {
			log.Warn("failed to close client connection", "error", err)
		}
		delete(s.connectedClients, fd)
	}

	log.Info("cleaned up all client connections")
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
		s.queryWatcher.Run(watchCtx)
	}()

	shardManagerCtx, cancelShardManager := context.WithCancel(ctx)
	defer cancelShardManager()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.shardManager.Run(shardManagerCtx)
	}()

	s.shardManager.RegisterWorker("server", s.ioChan)

	if err := syscall.Listen(s.serverFD, s.maxClients); err != nil {
		return err
	}

	var err error
	s.multiplexer, err = iomultiplexer.New(s.maxClients)
	if err != nil {
		return err
	}

	defer func() {
		if err := s.multiplexer.Close(); err != nil {
			log.Warn("failed to close multiplexer", "error", err)
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
			s.shardManager.UnregisterWorker("server")
			cancelEventLoop()
			cancelShardManager()
			cancelWatch()
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
						log.Warn(err)
					}
				} else {
					if err := s.handleClientEvent(event); err != nil {
						if errors.Is(err, ErrAborted) {
							log.Info("Received abort command, initiating graceful shutdown")
							return err
						}
						log.Warn(err)
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
			log.Printf("error closing client connection: %v", err)
		}
		delete(s.connectedClients, event.Fd)
		return err
	}

	s.EvalAndRespond(commands, client)
	if hasAbort {
		return ErrAborted
	}

	return nil
}

func (s *AsyncServer) executeCommandToBuffer(redisCmd *cmd.RedisCmd, buf *bytes.Buffer, c *comm.Client) {
	s.shardManager.GetShard(0).ReqChan <- &ops.StoreOp{
		Cmd:      redisCmd,
		WorkerID: "server",
		ShardID:  0,
		Client:   c,
	}

	response := <-s.ioChan
	buf.Write(response.Result)
}

func (s *AsyncServer) EvalAndRespond(cmds cmd.RedisCmds, c *comm.Client) {
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, redisCmd := range cmds {
		if !s.isAuthenticated(redisCmd, c, buf) {
			continue
		}

		if c.IsTxn {
			s.handleTransactionCommand(redisCmd, c, buf)
		} else {
			s.handleNonTransactionCommand(redisCmd, c, buf)
		}
	}

	s.writeResponse(c, buf)
}

func (s *AsyncServer) isAuthenticated(redisCmd *cmd.RedisCmd, c *comm.Client, buf *bytes.Buffer) bool {
	if redisCmd.Cmd != auth.AuthCmd && !c.Session.IsActive() {
		buf.Write(core.Encode(errors.New("NOAUTH Authentication required"), false))
		return false
	}
	return true
}

func (s *AsyncServer) handleTransactionCommand(redisCmd *cmd.RedisCmd, c *comm.Client, buf *bytes.Buffer) {
	if core.TxnCommands[redisCmd.Cmd] {
		switch redisCmd.Cmd {
		case core.ExecCmdMeta.Name:
			s.executeTransaction(c, buf)
		case core.DiscardCmdMeta.Name:
			s.discardTransaction(c, buf)
		default:
			log.Errorf("Unhandled transaction command: %s", redisCmd.Cmd)
		}
	} else {
		c.TxnQueue(redisCmd)
		buf.Write(core.RespQueued)
	}
}

func (s *AsyncServer) handleNonTransactionCommand(redisCmd *cmd.RedisCmd, c *comm.Client, buf *bytes.Buffer) {
	switch redisCmd.Cmd {
	case core.MultiCmdMeta.Name:
		c.TxnBegin()
		buf.Write(core.RespOK)
	case core.ExecCmdMeta.Name:
		buf.Write(diceerrors.NewErrWithMessage("EXEC without MULTI"))
	case core.DiscardCmdMeta.Name:
		buf.Write(diceerrors.NewErrWithMessage("DISCARD without MULTI"))
	default:
		s.executeCommandToBuffer(redisCmd, buf, c)
	}
}

func (s *AsyncServer) executeTransaction(c *comm.Client, buf *bytes.Buffer) {
	_, err := fmt.Fprintf(buf, "*%d\r\n", len(c.Cqueue))
	if err != nil {
		log.Errorf("Error writing to buffer: %v", err)
		return
	}

	for _, cmd := range c.Cqueue {
		s.executeCommandToBuffer(cmd, buf, c)
	}

	c.Cqueue = make(cmd.RedisCmds, 0)
	c.IsTxn = false
}

func (s *AsyncServer) discardTransaction(c *comm.Client, buf *bytes.Buffer) {
	c.TxnDiscard()
	buf.Write(core.RespOK)
}

func (s *AsyncServer) writeResponse(c *comm.Client, buf *bytes.Buffer) {
	if _, err := c.Write(buf.Bytes()); err != nil {
		log.Error(err)
	}
}
