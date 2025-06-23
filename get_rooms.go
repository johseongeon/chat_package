package pkg

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (adder *UserManager) GetRooms(c *Client) []string {
	adder.Mu.Lock()
	defer adder.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := adder.Client.Database("ChatDB").Collection("users")

	filter := map[string]interface{}{"username": c.Username}
	projection := map[string]interface{}{
		"rooms": 1,
	}

	var result struct {
		Rooms []string `bson:"rooms"`
	}

	err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		log.Println("Error getting rooms:", err)
		return nil
	}

	return result.Rooms
}

func GetRooms(client *mongo.Client, userManager *UserManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(w)
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &Client{Username: username}
		rooms := userManager.GetRooms(clientObj)
		if rooms == nil {
			http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"rooms":    rooms,
		})
	}
}
