// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dicedb/dicedb-go"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/shardmanager"
	"github.com/dicedb/dice/internal/wal"
	"github.com/dicedb/dicedb-go/wire"
)

type IOThread struct {
	ClientID   string
	Mode       string
	Session    *auth.Session
	serverWire *dicedb.ServerWire
}

func NewIOThread(clientFD int) (*IOThread, error) {
	w, err := dicedb.NewServerWire(config.MaxRequestSize, config.KeepAlive, clientFD)
	if err != nil {
		if err.Kind == wire.NotEstablished {
			slog.Error("failed to establish connection to client", slog.Int("client-fd", clientFD), slog.Any("error", err))

			return nil, err.Unwrap()
		}
		slog.Error("unexpected error during client connection establishment, this should be reported to DiceDB maintainers", slog.Int("client-fd", clientFD))
		return nil, err.Unwrap()
	}

	return &IOThread{
		serverWire: w,
		Session:    auth.NewSession(),
	}, nil
}

func (t *IOThread) Start(ctx context.Context, shardManager *shardmanager.ShardManager, watchManager *WatchManager) error {
	for {
		var c *wire.Command
		recvCh := make(chan *wire.Command, 1)
		errCh := make(chan error, 1)

		go func() {
			tmpC, err := t.serverWire.Receive()
			if err != nil {
				errCh <- err.Unwrap()
				return
			}
			recvCh <- tmpC
		}()

		select {
		case <-ctx.Done():
			slog.Debug("io-thread context canceled, shutting down receive loop")
			return ctx.Err()
		case err := <-errCh:
			return err
		case tmp := <-recvCh:
			c = tmp
		}

		_c := &cmd.Cmd{
			C:        c,
			ClientID: t.ClientID,
			Mode:     t.Mode,
		}

		res, err := _c.Execute(shardManager)
		if err != nil {
			res = &cmd.CmdRes{
				Rs: &wire.Result{
					Status:  wire.Status_ERR,
					Message: err.Error(),
				},
			}
			if sendErr := t.serverWire.Send(ctx, res.Rs); sendErr != nil {
				return sendErr.Unwrap()
			}
			// Continue in case of error
			continue
		}

		res.Rs.Status = wire.Status_OK
		if res.Rs.Message == "" {
			res.Rs.Message = "OK"
		}

		// Log command to WAL if enabled and not a replay
		if wal.DefaultWAL != nil && !_c.IsReplay {
			if err := wal.DefaultWAL.LogCommand(_c.C); err != nil {
				slog.Error("failed to log command to WAL", slog.Any("error", err))
			}
		}

		// TODO: Optimize this. We are doing this for all command execution
		// Also, we are allowing people to override the client ID.
		// Also, CLientID is duplicated in command and io-thread.
		// Also, we shouldn't allow execution/registration incase of invalid commands
		// like for B.WATCH cmd since it'll err out we shall return and not create subscription
		if err == nil {
			t.ClientID = _c.ClientID
		}

		if _c.Meta.IsWatchable {
			_cWatch := _c
			_cWatch.C.Cmd += ".WATCH"
			res.Rs.Fingerprint64 = _cWatch.Fingerprint()
		}

		if c.Cmd == "HANDSHAKE" && err == nil {
			t.ClientID = _c.C.Args[0]
			t.Mode = _c.C.Args[1]
		}

		isWatchCmd := strings.HasSuffix(c.Cmd, "WATCH")

		if isWatchCmd {
			watchManager.HandleWatch(_c, t)
		} else if strings.HasSuffix(c.Cmd, "UNWATCH") {
			watchManager.HandleUnwatch(_c, t)
		}

		watchManager.RegisterThread(t)

		// Only send the response directly if this is not a watch command
		// For watch commands, the response will be sent by NotifyWatchers
		if !isWatchCmd {
			if sendErr := t.serverWire.Send(ctx, res.Rs); sendErr != nil {
				return sendErr.Unwrap()
			}
		}

		// TODO: Streamline this because we need ordering of updates
		// that are being sent to watchers.
		if err == nil {
			watchManager.NotifyWatchers(_c, shardManager, t)
		}
	}
}

func (t *IOThread) Stop() error {
	t.serverWire.Close()
	t.Session.Expire()
	return nil
}
