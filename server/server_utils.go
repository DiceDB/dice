package server

import (
	"context"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/internal/constants"
)

// Waits on `core.WatchChannel` to receive updates about keys. Sends the update
// to all the clients that are watching the key.
// The message sent to the client will contain the new value and the operation
// that was performed on the key.
func WatchKeys(ctx context.Context, wg *sync.WaitGroup, store *core.Store) {
	defer wg.Done()
	for {
		select {
		case event := <-core.WatchChannel:
			store.WatchList.Range(func(key, value interface{}) bool {
				query := key.(core.DSQLQuery)
				clients := value.(*sync.Map)

				if core.WildCardMatch(query.KeyRegex, event.Key) {
					queryResult, err := core.ExecuteQuery(query, store)
					if err != nil {
						log.Error(err)
						return true // continue to next item
					}

					encodedResult := core.Encode(core.CreatePushResponse(&query, &queryResult), false)
					clients.Range(func(clientKey, _ interface{}) bool {
						clientFd := clientKey.(int)
						_, err := syscall.Write(clientFd, encodedResult)
						if err != nil {
							store.RemoveWatcher(query, clientFd)
						}
						return true
					})
				}
				return true
			})
		case <-ctx.Done():
			log.Info("Context closed")
			return
		}
	}
}

func WaitForSignal(wg *sync.WaitGroup, sigs chan os.Signal) {
	defer wg.Done()
	<-sigs

	// if server is busy continue to wait
	for atomic.LoadInt32(&eStatus) == EngineStatusBUSY { //nolint:revive
	}

	// CRITICAL TO HANDLE
	// We do not want server to ever go back to BUSY state
	// when control flow is here ->

	// immediately set the status to be SHUTTING DOWN
	// the only place where we can set the status to be SHUTTING DOWN
	atomic.StoreInt32(&eStatus, EngineStatusSHUTTINGDOWN)

	// if server is in any other state, initiate a shutdown
	core.Shutdown(asyncStore)
	os.Exit(0) //nolint:gocritic
}

func FindPortAndBind() (int, error) {
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return 0, err
	}

	if err := syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return 0, err
	}

	// Set the Socket operate in a non-blocking mode
	if err := syscall.SetNonblock(serverFD, true); err != nil {
		return 0, err
	}

	// Bind the IP and the port
	ip4 := net.ParseIP(config.Host)

	if err := syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return 0, err
	}

	return serverFD, nil
}

func SetupUsers() {
	var (
		user *core.User
		err  error
	)
	log.Info("setting up default user.", "password required", config.RequirePass != constants.EmptyStr)
	if user, err = core.UserStore.Add(core.DefaultUserName); err != nil {
		log.Fatal(err)
	}
	if err = user.SetPassword(config.RequirePass); err != nil {
		log.Fatal(err)
	}
}

func ReadCommands(c io.ReadWriter) (core.RedisCmds, bool, error) {
	var hasABORT bool = false
	rp := core.NewRESPParser(c)
	values, err := rp.DecodeMultiple()
	if err != nil {
		return nil, false, err
	}

	var cmds []*core.RedisCmd = make([]*core.RedisCmd, 0)
	for _, value := range values {
		tokens := toArrayString(value.([]interface{}))
		cmd := strings.ToUpper(tokens[0])
		cmds = append(cmds, &core.RedisCmd{
			Cmd:  cmd,
			Args: tokens[1:],
		})

		if cmd == "ABORT" {
			hasABORT = true
		}
	}
	return cmds, hasABORT, err
}

func toArrayString(ai []interface{}) []string {
	as := make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as
}

func respond(cmds core.RedisCmds, c *core.Client, store *core.Store) {
	core.EvalAndRespond(cmds, c, store)
}
