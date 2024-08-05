package tests

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/charmbracelet/log"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/server"
	"github.com/dicedb/dice/testutils"
	redis "github.com/dicedb/go-dice"
)

func getLocalConnection() net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", config.Port))
	if err != nil {
		panic(err)
	}
	return conn
}

// deleteTestKeys is a utility to delete a list of keys before running a test
func deleteTestKeys(keysToDelete []string) {
	for _, key := range keysToDelete {
		core.Del(key)
	}
}

func getLocalSdk() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(":%d", config.Port),

		DialTimeout:           10 * time.Second,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		ContextTimeoutEnabled: true,

		MaxRetries: -1,

		PoolSize:        10,
		PoolTimeout:     30 * time.Second,
		ConnMaxIdleTime: time.Minute,
	})
}

func fireCommand(conn net.Conn, cmd string) interface{} {
	var err error
	args := testutils.ParseCommand(cmd)
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
	args := testutils.ParseCommand(cmd)
	_, err = conn.Write(core.Encode(args, false))
	if err != nil {
		log.Fatalf("error %s while firing command: %s", err, cmd)
	}

	return core.NewRESPParser(conn)
}

func runTestServer(wg *sync.WaitGroup) {
	config.IOBufferLength = 16
	config.Port = 8739

	var totalRetries int = 100
	var serverFD int = 0
	var err error

	for i := 0; i < totalRetries; i++ {
		serverFD, err = server.FindPortAndBind()
		if err == nil {
			break
		}

		if err.Error() == "address already in use" {
			log.Infof("port %d already in use, trying another port", config.Port)
			config.Port += 1
		} else {
			panic(err)
		}
	}
	if serverFD == 0 {
		log.Fatalf("Tried %d times, could not find any port. Cannot start DiceDB. Please try after some time.", totalRetries)
		return
	}

	fmt.Println("starting the test server on port", config.Port)
	wg.Add(1)
	go server.RunAsyncTCPServer(serverFD, wg)
}
