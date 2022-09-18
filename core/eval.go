package core

import (
	"errors"
	"log"
	"net"
)

func evalPING(args []string, c net.Conn) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func EvalAndRespond(cmd *RedisCmd, c net.Conn) error {
	log.Println("comamnd:", cmd.Cmd)
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, c)
	default:
		return evalPING(cmd.Args, c)
	}
}
