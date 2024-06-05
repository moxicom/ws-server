package ws

import (
	"log"
	"sync"
)

type Hub struct {
	Clients    sync.Map
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
}

func NewHub() *Hub {
	return &Hub{
		Clients:    sync.Map{},
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Printf("New Client<%v> has been registered", client.ID)
			h.Clients.Store(client.ID, client)
		case client := <-h.Unregister:
			if _, ok := h.Clients.Load(client.ID); ok {
				log.Printf("Connection with Client<%v> has been closed", client.ID)
				h.Clients.Delete(client.ID)
				close(client.Send)
			}
		case msg := <-h.Broadcast:
			v, ok := h.Clients.Load(msg.ToID)
			if !ok {
				continue
			}

			targetClient := v.(*Client)

			select {
			case targetClient.Send <- msg:
			default:
				close(targetClient.Send)
				h.Clients.Delete(targetClient.ID)
				log.Printf("Connection with Client<%v> has been closed", targetClient.ID)
			}
		}
	}
}
