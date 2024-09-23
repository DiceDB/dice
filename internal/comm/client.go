package comm

import (
	"fmt"
	"io"
	"net/http"
	"syscall"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/cmd"
)

// HTTPResponseClientReadWriter wraps an http.ResponseWriter and implements ClientReadWriter.
// This is done to make it compatible with Client struct
type HTTPResponseClientReadWriter struct {
	http.ResponseWriter
}

type Client struct {
	ReadWriter         io.ReadWriter
	Fd                 int
	Cqueue             cmd.RedisCmds
	IsTxn              bool
	Session            *auth.Session
	ClientIdentifierID uint32
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

// Implement the Write method (inherited from http.ResponseWriter)
func (w *HTTPResponseClientReadWriter) Write(p []byte) (n int, err error) {
	return w.ResponseWriter.Write(p)
}

// Implement the Read method as a no-op for http clients.
func (w *HTTPResponseClientReadWriter) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("Read not supported for HTTPResponseClientReadWriter")
}

func (w *HTTPResponseClientReadWriter) Flush() error {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
		return nil
	}
	return fmt.Errorf("Flush not supported")
}

func NewClient(fd int) *Client {
	return &Client{
		Fd:      fd,
		Cqueue:  make(cmd.RedisCmds, 0),
		Session: auth.NewSession(),
	}
}

func NewHTTPClient(writer *HTTPResponseClientReadWriter, clientIdentifierID uint32) *Client {
	return &Client{
		Cqueue:             make(cmd.RedisCmds, 0),
		Session:            auth.NewSession(),
		ReadWriter:         writer,
		ClientIdentifierID: clientIdentifierID,
	}
}
