package comm

import (
	"io"
	"syscall"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/cmd"
)

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

func NewHTTPQwatchClient(qwatchResponseChan chan QwatchResponse, clientIdentifierID uint32) *Client {
	return &Client{
		Cqueue:                 make(cmd.RedisCmds, 0),
		Session:                auth.NewSession(),
		ClientIdentifierID:     clientIdentifierID,
		HTTPQwatchResponseChan: qwatchResponseChan,
	}
}
