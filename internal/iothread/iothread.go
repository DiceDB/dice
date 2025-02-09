// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package iothread

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/clientio/iohandler"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/wire"
)

type IOThread struct {
	id                string
	IoHandler         iohandler.IOHandler
	Session           *auth.Session
	ioThreadReadChan  chan []byte      // Channel to send data to the command handler
	ioThreadWriteChan chan interface{} // Channel to receive data from the command handler
	ioThreadErrChan   chan error       // Channel to receive errors from the ioHandler
}

func NewIOThread(id string, ioHandler iohandler.IOHandler,
	ioThreadReadChan chan []byte, ioThreadWriteChan chan interface{},
	ioThreadErrChan chan error) *IOThread {
	return &IOThread{
		id:                id,
		IoHandler:         ioHandler,
		Session:           auth.NewSession(),
		ioThreadReadChan:  ioThreadReadChan,
		ioThreadWriteChan: ioThreadWriteChan,
		ioThreadErrChan:   ioThreadErrChan,
	}
}

func (t *IOThread) ID() string {
	return t.id
}

func (t *IOThread) Start(ctx context.Context) error {
	// local channels to communicate between Start and startInputReader goroutine
	incomingDataChan := make(chan []byte) // data channel
	readErrChan := make(chan error)       // error channel

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
		case data := <-incomingDataChan:
			t.ioThreadReadChan <- data
		case err := <-readErrChan:
			slog.Debug("Read error in io-thread, connection closed possibly", slog.String("id", t.id), slog.Any("error", err))
			t.ioThreadErrChan <- err
			return err
		case resp := <-t.ioThreadWriteChan:
			err := t.IoHandler.Write(ctx, resp)
			if err != nil {
				slog.Debug("error while sending response to the client", slog.String("id", t.id), slog.Any("error", err))
				continue
			}
			slog.Debug("wrote response to client", slog.Any("resp", resp))
		}
	}
}

func (t *IOThread) StartSync(
	ctx context.Context, execute func(c *cmd.Cmd) (*cmd.CmdRes, error),
	handleWatch func(c *cmd.Cmd, t *IOThread),
	handleUnwatch func(c *cmd.Cmd, t *IOThread),
	notifyWatchers func(c *cmd.Cmd, execute func(c *cmd.Cmd) (*cmd.CmdRes, error))) error {
	slog.Debug("io thread started", slog.String("id", t.id))
	for {
		c, err := t.IoHandler.ReadSync()
		if err != nil {
			return err
		}
		c.ThreadID = t.id
		res, err := execute(c)
		
		if err != nil {
			res = &cmd.CmdRes{R: &wire.Response{Err: err.Error()}}
		}
		
		if strings.HasSuffix(c.C.Cmd, ".WATCH") {
			handleWatch(c, t)
		}

		if strings.HasSuffix(c.C.Cmd, "UNWATCH") {
			handleUnwatch(c, t)
		}

		err = t.IoHandler.WriteSync(ctx, res)
		
		if err != nil {
			return err
		}

		go notifyWatchers(c, execute)
	}
}

// startInputReader continuously reads input data from the ioHandler and sends it to the incomingDataChan.
func (t *IOThread) startInputReader(ctx context.Context, incomingDataChan chan []byte, readErrChan chan error) {
	defer close(incomingDataChan)
	defer close(readErrChan)

	for {
		data, err := t.IoHandler.Read(ctx)
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

func (t *IOThread) Stop() error {
	t.Session.Expire()
	return nil
}
