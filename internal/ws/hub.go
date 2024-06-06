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
	wg         sync.WaitGroup
}

func NewHub() *Hub {
	return &Hub{
		Clients:    sync.Map{},
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
		wg:         sync.WaitGroup{},
	}
}

// Run accepts register in channels;
// accepts unregister in channels;
// send a message to client's <send> channel;
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Printf("New Client<%v> has been registered", client.ID)
			h.Clients.Store(client.ID, client)
			h.wg.Add(1)
		case client := <-h.Unregister:
			if _, ok := h.Clients.Load(client.ID); ok {
				h.Clients.Delete(client.ID)
				close(client.Send)
				log.Printf("Client<%v> has been unregistered", client.ID)
				h.wg.Done()
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
				h.Unregister <- targetClient
				targetClient.Con.Close()
			}
		}
	}
}

func (h *Hub) Shutdown() {
	h.Clients.Range(func(key, value interface{}) bool {
		func() {
			client := value.(*Client)
			client.Con.Close()
			h.Unregister <- client
		}()
		return true
	})
	h.wg.Wait()
	log.Println("Hub stopped")
}
