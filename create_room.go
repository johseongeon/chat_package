package chat_package

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (rm *RoomManager) CreateRoom(roomID string) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	room := &ChatRoom{
		Id:      roomID,
		Clients: make(map[*Client]bool),
	}
	rm.Rooms[roomID] = room

	// Update the MongoDB collection to add the room
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := rm.Client.Database("ChatDB").Collection("rooms")
	filter := map[string]interface{}{"room_id": roomID}
	update := map[string]interface{}{
		"$setOnInsert": map[string]interface{}{
			"room_id": roomID,
			"clients": []string{},
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error creating room:", err)
		return
	}
}

func CreateRoom(client *mongo.Client, rm *RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(w)
		roomID := r.URL.Query().Get("room_id")
		rm.CreateRoom(roomID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("created room successfully"))
	}
}
