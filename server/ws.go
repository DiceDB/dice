package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/gorilla/websocket"
)

type WSServer struct {
	server           *http.Server
	connectedClients map[int]*core.Client
	Store            *core.Store
}

// NewWSServer initializes a new WSServer
func NewWSServer(store *core.Store) *WSServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.WSPort),
		Handler: mux,
	}
	wsSrv := &WSServer{
		server:           srv,
		connectedClients: make(map[int]*core.Client),
		Store:            store,
	}
	mux.HandleFunc("/", wsSrv.WSHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	return wsSrv
}

// Run starts the WSServer, accepts connections, and handles client requests
func (s *WSServer) Run(ctx context.Context) error {
	_, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()

	log.Infof("WS server listening on port %d\n", config.WSPort)
	go func() {
		<-ctx.Done()
		s.server.Shutdown(ctx)
	}()
	return s.server.ListenAndServe()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *WSServer) WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	client := core.NewClient(-1, conn)
	for {
		commands, hasAbort, err := readCommands(conn.UnderlyingConn())
		if err != nil {
			log.Error(err)
			continue
		}
		fmt.Println(commands, hasAbort, err)

		// TODO: handle abort
		if hasAbort {
			log.Warn("abort invoked over WS", hasAbort)
		}

		// respond(commands, client, s.Store)
		log.Info(commands, client)

		err = conn.WriteMessage(websocket.TextMessage, []byte("ok"))
		if err != nil {
			log.Error(err)
			break
		}
	}
}
