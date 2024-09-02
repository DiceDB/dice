package comm

import (
	"io"
	"syscall"

	"github.com/dicedb/dice/core/auth"
	"github.com/dicedb/dice/core/cmd"
)

type Client struct {
	io.ReadWriter
	Fd      int
	Cqueue  cmd.RedisCmds
	IsTxn   bool
	Session *auth.Session
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
	c.Cqueue = make(cmd.RedisCmds, 0)
	c.IsTxn = false
}

func (c *Client) TxnQueue(redisCmd *cmd.RedisCmd) {
	c.Cqueue = append(c.Cqueue, redisCmd)
}

func NewClient(fd int) *Client {
	return &Client{
		Fd:      fd,
		Cqueue:  make(cmd.RedisCmds, 0),
		Session: auth.NewSession(),
	}
}
