package ws

import (
	"context"
	"github.com/moxicom/ws-server/api/internal/storage/redis_storage"
	"log/slog"
	"strconv"
	"sync"
)

const brockerChanPrefix = "msg_user_"

type Hub struct {
	Logger       *slog.Logger
	Clients      sync.Map
	Register     chan *Client
	Unregister   chan *Client
	FromPub      chan Message
	ToSub        chan Message
	wg           sync.WaitGroup
	RedisStorage *redis_storage.RedisService
}

func NewHub(logger *slog.Logger, rs *redis_storage.RedisService) *Hub {
	return &Hub{
		Logger:       logger,
		RedisStorage: rs,
		Clients:      sync.Map{},
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		FromPub:      make(chan Message),
		ToSub:        make(chan Message),

		wg: sync.WaitGroup{},
	}
}

// Run accepts register in channels;
// accepts unregister in channels;
// send a message to client's <send> channel;
func (h *Hub) Run() {
	logger := h.Logger.With(slog.String("op", "ws.Hub.Run"))

	for {
		select {
		case client := <-h.Register:
			logger.Info("New Client has been registered")
			h.Clients.Store(client.ID, client)
			h.wg.Add(1)

		case client := <-h.Unregister:
			if _, ok := h.Clients.Load(client.ID); ok {
				close(client.Send)
				h.Clients.Delete(client.ID)
				logger.Info("Client has been unregistered")
				h.wg.Done()
			}
		case msg := <-h.FromPub:
			h.RedisStorage.Publish(context.TODO(), brockerChanPrefix+strconv.FormatUint(uint64(msg.ToID), 10), msg)
		case msg := <-h.ToSub:
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
	logger := h.Logger.With(slog.String("op", "ws.Hub.Shutdown"))

	h.Clients.Range(func(key, value interface{}) bool {
		func() {
			client := value.(*Client)
			client.Con.Close()
			h.Unregister <- client
		}()
		return true
	})
	h.wg.Wait()
	logger.Info("Hub stopped")
}
