package core

import (
	"io"
	"syscall"
)

type Client struct {
	io.ReadWriter
	Fd      int
	Cqueue  RedisCmds
	IsTxn   bool
	Session *Session
}

func (c Client) Write(b []byte) (int, error) {
	return syscall.Write(c.Fd, b)
}

func (c Client) Read(b []byte) (int, error) {
	return syscall.Read(c.Fd, b)
}

func (c *Client) TxnBegin() {
	c.IsTxn = true
}

func (c *Client) TxnDiscard() {
	c.Cqueue = make(RedisCmds, 0)
	c.IsTxn = false
}

func (c *Client) TxnQueue(cmd *RedisCmd) {
	c.Cqueue = append(c.Cqueue, cmd)
}

func NewClient(fd int) *Client {
	return &Client{
		Fd:      fd,
		Cqueue:  make(RedisCmds, 0),
		Session: NewSession(),
	}
}
