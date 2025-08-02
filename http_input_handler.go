package main

import (
	"context"
	"log"
	"net/http"

	_ "embed"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

//go:embed index.html
var frontend []byte

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

type HttpInputHandler struct {
	playerService PlayerService
	server        *http.Server
	mux           *http.ServeMux
	playerIdMap   map[uuid.UUID]int
}

func NewHttpInputHandler(ps PlayerService) *HttpInputHandler {
	handler := &HttpInputHandler{
		playerService: ps,
		playerIdMap:   map[uuid.UUID]int{},

		mux: http.NewServeMux(),
	}

	handler.mux.HandleFunc("GET /", handler.serveFrontend)
	handler.mux.HandleFunc("GET /ws", handler.serveWebsocket)

	return handler
}

func (k *HttpInputHandler) Listen(addr string) {
	k.server = &http.Server{
		Addr:    addr,
		Handler: k.mux,
	}
	k.server.ListenAndServe()
}

func (h *HttpInputHandler) serveFrontend(w http.ResponseWriter, req *http.Request) {
	w.Write(frontend)
}

func (h *HttpInputHandler) serveWebsocket(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	playerId := uuid.New()
	h.playerService.Join(playerId)

	for {
		// Read a message from the client
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		switch string(message) {
		case "left":
			h.playerService.TurnLeft(playerId)
		case "right":
			h.playerService.TurnRight(playerId)
		}
		conn.WriteMessage(websocket.TextMessage, []byte("response"))
	}
}

func (k *HttpInputHandler) Close() {
	if k.server == nil {
		return
	}
	k.server.Shutdown(context.Background())
}
