package core

import (
	"bytes"
	"fmt"
	"io"
	"syscall"
)

type Client struct {
	io.ReadWriter
	Fd     int
	cqueue RedisCmds
	isTxn  bool
}

func (c Client) Write(b []byte) (int, error) {
	return syscall.Write(c.Fd, b)
}

func (c Client) Read(b []byte) (int, error) {
	return syscall.Read(c.Fd, b)
}

func (c *Client) TxnBegin() {
	c.isTxn = true
}

func (c *Client) TxnExec() []byte {
	var out []byte
	buf := bytes.NewBuffer(out)

	buf.WriteString(fmt.Sprintf("*%d\r\n", len(c.cqueue)))
	for _, _cmd := range c.cqueue {
		buf.Write(executeCommand(_cmd, c))
	}

	c.cqueue = make(RedisCmds, 0)
	c.isTxn = false

	return buf.Bytes()
}

func (c *Client) TxnDiscard() {
	c.cqueue = make(RedisCmds, 0)
	c.isTxn = false
}

func (c *Client) TxnQueue(cmd *RedisCmd) {
	c.cqueue = append(c.cqueue, cmd)
}

func NewClient(fd int) *Client {
	return &Client{
		Fd:     fd,
		cqueue: make(RedisCmds, 0),
	}
}
