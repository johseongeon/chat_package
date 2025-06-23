package chat_package

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func (rm *RoomManager) JoinRoom(c *Client, room *ChatRoom) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	room.Mu.Lock()
	defer room.Mu.Unlock()

	c.Rooms[room.Id] = room
	room.Clients[c] = true

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := rm.Client.Database("ChatDB").Collection("users")

	userFilter := map[string]interface{}{"username": c.Username}
	userUpdate := map[string]interface{}{
		"$addToSet": map[string]interface{}{
			"rooms": room.Id,
		},
	}

	_, err := userCollection.UpdateOne(ctx, userFilter, userUpdate)
	if err != nil {
		log.Println("Error updating user's rooms in 'users' collection:", err)
		return
	}

	roomCollection := rm.Client.Database("ChatDB").Collection("rooms")

	roomFilter := map[string]interface{}{"room_id": room.Id}
	roomUpdate := map[string]interface{}{
		"$addToSet": map[string]interface{}{
			"clients": c.Username,
		},
	}

	_, err = roomCollection.UpdateOne(ctx, roomFilter, roomUpdate)
	if err != nil {
		log.Println("Error updating room's clients in 'rooms' collection:", err)
	}
}

func JoinUser(client *mongo.Client, rm *RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(w)
		username := r.URL.Query().Get("username")
		roomID := r.URL.Query().Get("room_id")
		clientObj := &Client{Username: username, Rooms: make(map[string]*ChatRoom)}
		room := RoomMgr.GetRoom(roomID)
		RoomMgr.JoinRoom(clientObj, room)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Joined room successfully"))
	}
}
