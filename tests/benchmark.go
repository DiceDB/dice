package tests

import (
	"io"
	"log"
	"net"
	"strings"

	"github.com/dicedb/dice/core"
)

func getLocalConnection() net.Conn {
	conn, err := net.Dial("tcp", "localhost:8379")
	if err != nil {
		panic(err)
	}
	return conn
}

func fireCommand(conn net.Conn, cmd string) interface{} {
	var err error
	_, err = conn.Write(core.Encode(strings.Split(cmd, " "), false))
	if err != nil {
		log.Fatalf("error %s while firing command: %s", err, cmd)
	}

	rp := core.NewRESPParser(conn)
	v, err := rp.DecodeOne()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		log.Fatalf("error %s while firing command: %s", err, cmd)
	}
	return v
}
