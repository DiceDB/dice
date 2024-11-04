package comm

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dicedb/dice/internal/auth"
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/id"
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
}

func (c *Client) String() string {
	var s strings.Builder

	// id
	s.WriteString("id=")
	s.WriteString(strconv.FormatUint(uint64(c.ClientIdentifierID), 10))
	s.WriteString(" ")

	// addr
	s.WriteString("addr=")
	sa, err := syscall.Getpeername(c.Fd)
	if err != nil {
		return ""
	}
	var addr string
	switch v := sa.(type) {
	case *syscall.SockaddrInet4:
		addr = net.IP(v.Addr[:]).String() + ":" + strconv.Itoa(v.Port)
	case *syscall.SockaddrInet6:
		addr = net.IP(v.Addr[:]).String() + ":" + strconv.Itoa(v.Port)
	}
	s.WriteString(addr)
	s.WriteString(" ")

	// laddr
	s.WriteString("laddr=")
	sa, err = syscall.Getsockname(c.Fd)
	if err != nil {
		return ""
	}
	switch v := sa.(type) {
	case *syscall.SockaddrInet4:
		addr = net.IP(v.Addr[:]).String() + ":" + strconv.Itoa(v.Port)
	case *syscall.SockaddrInet6:
		addr = net.IP(v.Addr[:]).String() + ":" + strconv.Itoa(v.Port)
	}
	s.WriteString(addr)
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

	// cmd
	s.WriteString("cmd=")
	// todo: handle `CLIENT ID` as "client|id" and `SET k 1` as "set"
	s.WriteString(c.LastCmd.Cmd)
	s.WriteString(" ")

	// this breaks
	// user
	// s.WriteString("user=")
	// s.WriteString(c.Session.User.Username)
	// s.WriteString(" ")
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

	addr, err := syscall.Getsockname(fd)
	fmt.Println("addr: ", addr, err)
	switch v := addr.(type) {
	case *syscall.SockaddrInet4:
		fmt.Println("addr: ", v.Addr, v.Port)
	case *syscall.SockaddrInet6:
		fmt.Println("addr: ", v.Addr, v.Port, v.ZoneId)
	}

	// remoteAddr, err := syscall.Getpeername(fd)
	// fmt.Println("addr: ", remoteAddr, err)

	return &Client{
		Fd: fd,
		Cqueue: cmd.RedisCmds{
			Cmds: cmds,
		},
		Session:            auth.NewSession(),
		ClientIdentifierID: uint32(id.NextClientID()), // this should be int64 as per redis
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
