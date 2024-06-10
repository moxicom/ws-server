package ws

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
)

var upgrader = websocket.Upgrader{
	// Reuse buffers that allocates standard library http server
	ReadBufferSize:  0,
	WriteBufferSize: 0,
}

const UserContextKey = "user"

func Upgrade(h *Hub, w http.ResponseWriter, r *http.Request) {
	logger := h.Logger.With(slog.String("op", "ws.Upgrade"))

	logger.Debug("Handled WS")

	userID := r.Context().Value(UserContextKey).(uint64)

	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	fmt.Println(userID)

	client := &Client{ID: userID, Hub: h, Con: con, Send: make(chan Message, 256)}
	client.Hub.Register <- client

	ctx, cancel := context.WithCancel(context.Background())
	go client.readWS(cancel)
	go client.writeWS()
	go client.BrokerBroadcast(ctx, cancel, client.ID)
}
