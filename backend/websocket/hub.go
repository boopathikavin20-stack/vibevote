package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

type Client struct {
	Conn     *websocket.Conn
	PollID   string
	Send     chan []byte
	Hub      *Hub
	CloseNow func()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewHub() *Hub {
	return &Hub{clients: map[string]*Client{}}
}

func (h *Hub) RegisterClient(pollID string, conn *websocket.Conn) *Client {
	client := &Client{Conn: conn, PollID: pollID, Send: make(chan []byte, 10), Hub: h}
	client.CloseNow = func() {
		conn.Close()
		h.RemoveClient(client)
	}
	h.mu.Lock()
	h.clients[pollID+":"+conn.RemoteAddr().String()] = client
	h.mu.Unlock()
	return client
}

func (h *Hub) RemoveClient(client *Client) {
	h.mu.Lock()
	delete(h.clients, client.PollID+":"+client.Conn.RemoteAddr().String())
	h.mu.Unlock()
}

func (h *Hub) BroadcastPollUpdate(pollID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		if client.PollID == pollID {
			select {
			case client.Send <- payload:
			default:
				log.Println("websocket send buffer full")
			}
		}
	}
}

func ServePollWebSocket(c *gin.Context, hub *Hub, pollID string) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("websocket upgrade error:", err)
		return
	}
	client := hub.RegisterClient(pollID, conn)
	defer client.CloseNow()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
