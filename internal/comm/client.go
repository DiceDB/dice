// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
	Cqueue                 cmd.DiceDBCmds
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
		Cqueue: cmd.DiceDBCmds{
			Cmds: cmds,
		},
		Session: auth.NewSession(),
	}
}

func NewHTTPQwatchClient(qwatchResponseChan chan QwatchResponse, clientIdentifierID uint32) *Client {
	cmds := make([]*cmd.DiceDBCmd, 0)
	return &Client{
		Cqueue:                 cmd.DiceDBCmds{Cmds: cmds},
		Session:                auth.NewSession(),
		ClientIdentifierID:     clientIdentifierID,
		HTTPQwatchResponseChan: qwatchResponseChan,
	}
}
