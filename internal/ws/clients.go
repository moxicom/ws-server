package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 5 * time.Second
	pongWait       = 5 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	ID   uint64
	Hub  *Hub
	Con  *websocket.Conn
	Send chan Message
}

type Message struct {
	FromID uint64 `json:"from_id"`
	ToID   uint64 `json:"to_id,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

func (c *Client) readWS() {
	defer func() {
		c.Hub.Unregister <- c
		c.Con.Close()
	}()

	logger := c.Hub.Logger.With(slog.String("op", "ws.Client.readWS"))

	c.Con.SetReadLimit(maxMessageSize)
	c.Con.SetReadDeadline(time.Now().Add(pongWait))
	c.Con.SetPongHandler(func(appData string) error {
		c.Con.SetReadDeadline(time.Now().Add(pongWait))
		logger.Debug("Pong handler")
		return nil
	})

	for {
		_, byteMsg, err := c.Con.ReadMessage()
		logger.Debug("Accepted a message")

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("error: %v\n", err.Error())
			}
			break
		}

		var msg Message
		err = json.Unmarshal(byteMsg, &msg)
		if err != nil {
			logger.Error("error %v\n", err.Error())
			break
		}

		if msg.Msg == "" || msg.ToID == 0 || msg.FromID == 0 {
			logger.Error("error %v\n", "validation")
			break
		}
		c.Hub.Broadcast <- msg
	}
}

func (c *Client) writeWS() {
	ticker := time.NewTicker(pingPeriod)
	logger := c.Hub.Logger.With(slog.String("op", "ws.Client.writeWS"))

	defer func() {
		ticker.Stop()
		c.Con.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Con.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Con.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Con.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error(err.Error())
				return
			}

			msgJson, err := json.Marshal(msg)
			if err != nil {
				logger.Error(err.Error())
				return
			}

			w.Write(msgJson)

			if err := w.Close(); err != nil {
				logger.Error(err.Error())
				return
			}
		case <-ticker.C:
			c.Con.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Con.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error(err.Error())
				return
			}
		}
	}
}
