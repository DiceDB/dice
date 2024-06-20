package tests

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/server"
)

const (
	serverPort = 8379
)

func getLocalConnection() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", serverPort))
	if err != nil {
		panic(err)
	}
	return conn
}

func fireCommand(conn net.Conn, cmd string) interface{} {
	var err error
	args := parseCommand(cmd)
	_, err = conn.Write(core.Encode(args, false))
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

func fireCommandAndGetRESPParser(conn net.Conn, cmd string) *core.RESPParser {
	var err error
	args := parseCommand(cmd)
	_, err = conn.Write(core.Encode(args, false))
	if err != nil {
		log.Fatalf("error %s while firing command: %s", err, cmd)
	}

	return core.NewRESPParser(conn)
}

func runTestServer(wg *sync.WaitGroup) {
	config.IOBufferLength = 16
	config.Port = serverPort
	wg.Add(1)
	server.RunAsyncTCPServer(wg)
}

func parseCommand(cmd string) []string {
	var args []string
	var current string
	var inQuotes bool

	for _, char := range cmd {
		switch char {
		case ' ':
			if inQuotes {
				current += string(char)
			} else {
				if len(current) > 0 {
					args = append(args, current)
					current = ""
				}
			}
		case '"':
			inQuotes = !inQuotes
			current += string(char)
		default:
			current += string(char)
		}
	}

	if len(current) > 0 {
		args = append(args, current)
	}

	// Remove quotes from each argument
	for i, arg := range args {
		args[i] = strings.Trim(arg, `"`)
	}

	return args
}
