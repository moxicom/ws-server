package ws

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Reuse buffers that allocates standard library http server
	ReadBufferSize:  0,
	WriteBufferSize: 0,
}

func Upgrade(h *Hub, w http.ResponseWriter, r *http.Request) {
	logger := h.Logger.With(slog.String("op", "ws.Upgrade"))

	logger.Debug("Handled WS")

	userIDstr := r.URL.Query().Get("id") // or any data to auth a client
	if userIDstr == "" {
		http.Error(w, fmt.Errorf("no id in query").Error(), http.StatusBadRequest)
		logger.Error("error: %s\n", "No id in query")
		return
	}

	userID, err := strconv.ParseUint(userIDstr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Error(err.Error())
		return
	}

	//
	// MOCK VALIDATION
	//
	if false {
		http.Error(w, fmt.Errorf("unauthorized").Error(), http.StatusUnauthorized)
	}

	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	client := &Client{ID: userID, Hub: h, Con: con, Send: make(chan Message, 256)}
	client.Hub.Register <- client

	go client.readWS()
	go client.writeWS()
}
