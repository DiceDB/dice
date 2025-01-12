// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package comm

import (
	"io"
	"syscall"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/cmd"
)

type CmdWatchResponse struct {
	ClientIdentifierID uint32
	Result             interface{}
	Error              error
}

type QwatchResponse struct {
	ClientIdentifierID uint32
	Result             interface{}
	Error              error
}

type Client struct {
	io.ReadWriter
	HTTPQwatchResponseChan chan QwatchResponse // Response channel to send back the operation response
	Fd                     int
	Cqueue                 cmd.RedisCmds
	IsTxn                  bool
	Session                *auth.Session
	ClientIdentifierID     uint32
}

func (c *Client) Write(b []byte) (int, error) {
	return syscall.Write(c.Fd, b)
}

func (c *Client) Read(b []byte) (int, error) {
	return syscall.Read(c.Fd, b)
}

func (c *Client) TxnBegin() {
	c.IsTxn = true
}

func (c *Client) TxnDiscard() {
	c.Cqueue.Cmds = make([]*cmd.DiceDBCmd, 0)
	c.IsTxn = false
}

func (c *Client) TxnQueue(diceDBCmd *cmd.DiceDBCmd) {
	c.Cqueue.Cmds = append(c.Cqueue.Cmds, diceDBCmd)
}

func NewClient(fd int) *Client {
	cmds := make([]*cmd.DiceDBCmd, 0)
	return &Client{
		Fd: fd,
		Cqueue: cmd.RedisCmds{
			Cmds: cmds,
		},
		Session: auth.NewSession(),
	}
}

func NewHTTPQwatchClient(qwatchResponseChan chan QwatchResponse, clientIdentifierID uint32) *Client {
	cmds := make([]*cmd.DiceDBCmd, 0)
	return &Client{
		Cqueue:                 cmd.RedisCmds{Cmds: cmds},
		Session:                auth.NewSession(),
		ClientIdentifierID:     clientIdentifierID,
		HTTPQwatchResponseChan: qwatchResponseChan,
	}
}
