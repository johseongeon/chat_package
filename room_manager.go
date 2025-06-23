package chat_package

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// handles all chat room operations
type RoomManager struct {
	Rooms  map[string]*ChatRoom
	Mu     sync.RWMutex
	Client *mongo.Client
}

var RoomMgr = &RoomManager{
	Rooms:  make(map[string]*ChatRoom),
	Client: nil,
}

func (rm *RoomManager) GetRoom(roomID string) *ChatRoom {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	if room, exists := rm.Rooms[roomID]; exists {
		return room
	}

	return nil
}

func LoadWhileRunning(mgr *RoomManager) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	collection := mgr.Client.Database("ChatDB").Collection("rooms")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error loading rooms from DB: %v", err)
		return
	}
	defer cursor.Close(ctx)

	mgr.Mu.Lock()
	defer mgr.Mu.Unlock()

	for cursor.Next(ctx) {
		var roomDoc struct {
			RoomID  string   `bson:"room_id"`
			Clients []string `bson:"clients"`
		}

		err := cursor.Decode(&roomDoc)
		if err != nil {
			log.Printf("Error decoding room document: %v", err)
			continue
		}

		room, exists := mgr.Rooms[roomDoc.RoomID]
		if !exists {
			// 기존에 없던 새 room만 생성
			room = &ChatRoom{
				Id:      roomDoc.RoomID,
				Clients: make(map[*Client]bool),
			}
			mgr.Rooms[roomDoc.RoomID] = room
		}
	}
}

func LoadRoomsFromDB(mgr *RoomManager) {
	mgr.Mu.Lock()
	defer mgr.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := mgr.Client.Database("ChatDB").Collection("rooms")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error loading rooms from DB: %v", err)
		return
	}
	defer cursor.Close(ctx)

	mgr.Rooms = make(map[string]*ChatRoom) // initialize the map

	for cursor.Next(ctx) {
		var roomDoc struct {
			RoomID  string   `bson:"room_id"`
			Clients []string `bson:"clients"`
		}

		if err := cursor.Decode(&roomDoc); err != nil {
			log.Printf("Error decoding room document: %v", err)
			continue
		}

		mgr.Rooms[roomDoc.RoomID] = &ChatRoom{
			Id:      roomDoc.RoomID,
			Clients: make(map[*Client]bool),
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error after iteration: %v", err)
	}
}

func (rm *RoomManager) ConnectToRoom(client *Client, room *ChatRoom) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	if room == nil {
		log.Println("Room is nil, cannot connect.")
		return
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	if _, exists := room.Clients[client]; !exists {
		room.Clients[client] = true
		client.Rooms[room.Id] = room

		log.Printf("Client %s connected to room %s", client.Username, room.Id)
	} else {
		log.Printf("Client %s is already connected to room %s", client.Username, room.Id)
	}
}

func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()
	delete(rm.Rooms, roomID)

	// Update the MongoDB collection to remove the room
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	collection := rm.Client.Database("ChatDB").Collection("rooms")
	filter := bson.M{"room_id": roomID}
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println("Error removing room:", err)
		return
	}

}

func (c *Client) LeaveRoom(roomID string) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if room, exists := c.Rooms[roomID]; exists {
		room.Mu.Lock()
		delete(room.Clients, c)
		room.Mu.Unlock()

		delete(c.Rooms, roomID)

		room.Mu.RLock()
		if len(room.Clients) == 0 {
			room.Mu.RUnlock()
			RoomMgr.RemoveRoom(roomID)
		} else {
			room.Mu.RUnlock()
		}
	}
}
