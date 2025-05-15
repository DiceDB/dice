// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"syscall"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/shardmanager"
)

type Server struct {
	Host            string
	Port            int
	serverFD        int
	connBacklogSize int
	shardManager    *shardmanager.ShardManager
	watchManager    *WatchManager
	ioThreadManager *IOThreadManager
}

func NewServer(shardManager *shardmanager.ShardManager, ioThreadManager *IOThreadManager, watchManager *WatchManager) *Server {
	return &Server{
		Host:            config.Config.Host,
		Port:            config.Config.Port,
		connBacklogSize: config.DefaultConnBacklogSize,
		shardManager:    shardManager,
		ioThreadManager: ioThreadManager,
		watchManager:    watchManager,
	}
}

func (s *Server) Run(ctx context.Context) (err error) {
	if err = s.BindAndListen(); err != nil {
		slog.Error("failed to bind server", slog.Any("error", err))
		return err
	}

	defer releasePort(s.serverFD)

	errChan := make(chan error, 1)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := s.AcceptConnectionRequests(ctx, wg); err != nil {
			errChan <- fmt.Errorf("failed to accept connections %w", err)
		}
	}(wg)

	select {
	case <-ctx.Done():
		slog.Info("initiating shutdown")
	case err = <-errChan:
		slog.Error("error while accepting connections, initiating shutdown", slog.Any("error", err))
	}

	s.Shutdown()

	wg.Wait() // Wait for the go routines to finish
	slog.Info("exiting gracefully")

	return err
}

func (s *Server) BindAndListen() error {
	serverFD, socketErr := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if socketErr != nil {
		return fmt.Errorf("failed to create socket: %w", socketErr)
	}

	// Close the socket on exit if an error occurs
	var err error
	defer func() {
		if err != nil {
			if closeErr := syscall.Close(serverFD); closeErr != nil {
				// Wrap the close error with the original bind/listen error
				slog.Error("Error occurred", slog.Any("error", err), "additionally, failed to close socket", slog.Any("close-err", closeErr))
			} else {
				slog.Error("Error occurred", slog.Any("error", err))
			}
		}
	}()

	if err = syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return fmt.Errorf("failed to set SO_REUSEADDR: %w", err)
	}

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return fmt.Errorf("failed to set socket to non-blocking: %w", err)
	}

	ip4 := net.ParseIP(s.Host)
	if ip4 == nil {
		return fmt.Errorf("invalid IP address: %s", s.Host)
	}

	sockAddr := &syscall.SockaddrInet4{
		Port: s.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}
	if err = syscall.Bind(serverFD, sockAddr); err != nil {
		return fmt.Errorf("failed to bind socket: %w", err)
	}

	if err = syscall.Listen(serverFD, s.connBacklogSize); err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	s.serverFD = serverFD
	return nil
}

func releasePort(serverFD int) {
	if err := syscall.Close(serverFD); err != nil {
		slog.Error("Failed to close server socket", slog.Any("error", err))
	}
}

func (s *Server) AcceptConnectionRequests(ctx context.Context, wg *sync.WaitGroup) error {
	for {
		select {
		case <-ctx.Done():
			slog.Info("no new connections will be accepted")
			return ctx.Err()
		default:
			clientFD, _, err := syscall.Accept(s.serverFD)
			if err != nil {
				if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
					continue // No more connections to accept at this time
				}
				return fmt.Errorf("error accepting connection: %w", err)
			}

			thread, err := NewIOThread(clientFD)
			if err != nil {
				slog.Error("failed to create io-thread", slog.String("id", "-xxx"), slog.Any("error", err))
				continue
			}

			wg.Add(1)
			go s.startIOThread(ctx, wg, thread)
		}
	}
}

func (s *Server) startIOThread(ctx context.Context, wg *sync.WaitGroup, thread *IOThread) {
	defer wg.Done()
	err := thread.Start(ctx, s.shardManager, s.watchManager)
	if err != nil {
		if err == io.EOF {
			s.watchManager.CleanupThreadWatchSubscriptions(thread)
			slog.Debug("client disconnected. io-thread stopped",
				slog.String("client_id", thread.ClientID),
				slog.String("mode", thread.Mode),
			)
		} else {
			slog.Debug("io-thread errored out",
				slog.String("client_id", thread.ClientID),
				slog.String("mode", thread.Mode),
				slog.Any("error", err))
		}
	}
}

func (s *Server) Shutdown() {
	// Not implemented
}
