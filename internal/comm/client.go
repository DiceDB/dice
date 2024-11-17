package comm

import (
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/cmd"
)

const (
	flagTxn  = "x"
	flagNrml = "N"
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
	Name                   string
	Fd                     int
	Cqueue                 cmd.RedisCmds
	IsTxn                  bool
	Session                *auth.Session
	ClientIdentifierID     uint32
	LastCmd                *cmd.DiceDBCmd
	ArgLenSum              int
}

func (c *Client) flag() string {
	if c.IsTxn {
		return flagTxn
	}
	return flagNrml
}

func (c *Client) addr() (string, string, error) {
	// addr
	sa, err := syscall.Getpeername(c.Fd)
	if err != nil {
		return "", "", err
	}
	var addr string
	switch v := sa.(type) {
	case *syscall.SockaddrInet4:
		addr = net.IP(v.Addr[:]).String() + ":" + strconv.Itoa(v.Port)
	case *syscall.SockaddrInet6:
		addr = net.IP(v.Addr[:]).String() + ":" + strconv.Itoa(v.Port)
	}

	// laddr
	sa, err = syscall.Getsockname(c.Fd)
	if err != nil {
		return "", "", err
	}
	var laddr string
	switch v := sa.(type) {
	case *syscall.SockaddrInet4:
		laddr = net.IP(v.Addr[:]).String() + ":" + strconv.Itoa(v.Port)
	case *syscall.SockaddrInet6:
		laddr = net.IP(v.Addr[:]).String() + ":" + strconv.Itoa(v.Port)
	}

	return addr, laddr, nil
}

func (c *Client) String() string {
	var s strings.Builder

	// id
	s.WriteString("id=")
	s.WriteString(strconv.FormatUint(uint64(c.ClientIdentifierID), 10))
	s.WriteString(" ")

	// addr and laddr
	addr, laddr, err := c.addr()
	if err != nil {
		// log error?
		log.Println(err)
	}

	s.WriteString("addr=")
	s.WriteString(addr)
	s.WriteString(" ")

	s.WriteString("laddr=")
	s.WriteString(laddr)
	s.WriteString(" ")

	// fd
	s.WriteString("fd=")
	s.WriteString(strconv.FormatInt(int64(c.Fd), 10))
	s.WriteString(" ")

	// name
	s.WriteString("name=")
	s.WriteString(c.Name)
	s.WriteString(" ")

	// age
	s.WriteString("age=")
	s.WriteString(strconv.FormatFloat(time.Since(c.Session.CreatedAt).Seconds(), 'f', 0, 64))
	s.WriteString(" ")

	// idle
	s.WriteString("idle=")
	s.WriteString(strconv.FormatFloat(time.Since(c.Session.LastAccessedAt).Seconds(), 'f', 0, 64))
	s.WriteString(" ")

	// flags
	s.WriteString("flags=")
	s.WriteString(c.flag())
	s.WriteString(" ")

	// multi
	s.WriteString("multi=")
	multi := -1
	if c.IsTxn {
		multi = len(c.Cqueue.Cmds)
	}
	s.WriteString(strconv.FormatInt(int64(multi), 10))
	s.WriteString(" ")

	// cmd
	s.WriteString("cmd=")
	// todo: handle `CLIENT ID` as "client|id" and `SET k 1` as "set"
	s.WriteString(strings.ToLower(c.LastCmd.Cmd))
	s.WriteString(" ")

	// user
	s.WriteString("user=")
	if c.Session == nil || c.Session.User == nil {
		s.WriteString("default")
	} else {
		s.WriteString(c.Session.User.Username)
	}
	s.WriteString(" ")

	// argv-mem
	s.WriteString("argv-mem=")
	s.WriteString(strconv.FormatInt(int64(c.ArgLenSum), 10))
	s.WriteString(" ")

	return s.String()
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
		Session:            auth.NewSession(),
		ClientIdentifierID: uint32(nextClientID()), // this should be int64 as per redis
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
