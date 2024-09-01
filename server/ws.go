package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/dicedb/dice/config"
	"github.com/gorilla/websocket"
)

type WSServer struct {
	Server *http.Server
}

// NewWSServer initializes a new WSServer
func NewWSServer() *WSServer {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.WSPort),
		Handler: mux,
	}
	mux.HandleFunc("/", WSHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	return &WSServer{
		Server: srv,
	}
}

// Run starts the WSServer, accepts connections, and handles client requests
func (s *WSServer) Run(ctx context.Context) error {
	_, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()

	log.Infof("WS server listening on port %d\n", config.WSPort)
	go func() {
		<-ctx.Done()
		s.Server.Shutdown(ctx)
	}()
	return s.Server.ListenAndServe()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Error(err)
			break
		}

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Error(err)
			break
		}
	}
}
