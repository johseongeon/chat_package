package pkg

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Username string // identifier
	Conn     *websocket.Conn
	Rooms    map[string]*ChatRoom
	Friends  map[string]*Client
	Mu       sync.RWMutex
}

type ChatRoom struct {
	Id      string // identifier
	Clients map[*Client]bool
	Mu      sync.RWMutex
}

type ChatMessage struct {
	Username  string    `json:"username" bson:"username"`
	Message   string    `json:"message" bson:"message"`
	RoomID    string    `json:"room_id" bson:"room_id"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

func EnableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func (c *Client) BroadcastToRoom(roomID string, message map[string]string) {
	c.Mu.RLock()
	room, exists := c.Rooms[roomID]
	c.Mu.RUnlock()

	if !exists {
		return
	}

	room.Mu.RLock()
	defer room.Mu.RUnlock()

	for client := range room.Clients {
		if client != c {
			client.Conn.WriteJSON(message)
		}
	}
}
