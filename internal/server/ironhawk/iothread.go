// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/shardmanager"
	"github.com/dicedb/dicedb-go/wire"
)

type IOThread struct {
	ClientID  string
	Mode      string
	IoHandler *IOHandler
	Session   *auth.Session
}

func NewIOThread(clientFD int) (*IOThread, error) {
	io, err := NewIOHandler(clientFD)
	if err != nil {
		slog.Error("Failed to create new IOHandler for clientFD", slog.Int("client-fd", clientFD), slog.Any("error", err))
		return nil, err
	}
	return &IOThread{
		IoHandler: io,
		Session:   auth.NewSession(),
	}, nil
}

func (t *IOThread) StartSync(ctx context.Context, shardManager *shardmanager.ShardManager, watchManager *WatchManager) error {
	for {
		c, err := t.IoHandler.ReadSync()
		if err != nil {
			return err
		}

		_c := &cmd.Cmd{
			C:        c,
			ClientID: t.ClientID,
			Mode:     t.Mode,
		}

		res, err := _c.Execute(shardManager)
		if err != nil {
			res = &cmd.CmdRes{R: &wire.Response{CorrId: c.GetCorrId(), Err: err.Error()}}
		}

		// TODO: Optimize this. We are doing this for all command execution
		// Also, we are allowing people to override the client ID.
		// Also, CLientID is duplicated in command and io-thread.
		// Also, we shouldn't allow execution/registration incase of invalid commands
		// like for B.WATCH cmd since it'll err out we shall return and not create subscription
		t.ClientID = _c.ClientID

		if c.Cmd == "HANDSHAKE" {
			t.ClientID = _c.C.Args[0]
			t.Mode = _c.C.Args[1]
		}

		if strings.HasSuffix(c.Cmd, ".WATCH") {
			watchManager.HandleWatch(_c, t)
		}

		if strings.HasSuffix(c.Cmd, "UNWATCH") {
			watchManager.HandleUnwatch(_c, t)
		}

		watchManager.RegisterThread(t)

		if err := t.IoHandler.WriteSync(ctx, res.R); err != nil {
			return err
		}

		// TODO: Streamline this because we need ordering of updates
		// that are being sent to watchers.
		watchManager.NotifyWatchers(_c, shardManager, t)
	}
}

func (t *IOThread) Stop() error {
	t.Session.Expire()
	return nil
}
