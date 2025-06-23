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

func (adder *UserManager) GetFriends(c *Client) []string {
	adder.Mu.Lock()
	defer adder.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := adder.Client.Database("ChatDB").Collection("users")

	filter := map[string]interface{}{"username": c.Username}
	projection := map[string]interface{}{
		"friends": 1,
	}

	var result struct {
		Friends []string `bson:"friends"`
	}

	err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		log.Println("Error getting friends:", err)
		return nil
	}

	return result.Friends
}

func GetFriends(client *mongo.Client, userManager *UserManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(w)
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &Client{Username: username}
		friends := userManager.GetFriends(clientObj)
		if friends == nil {
			http.Error(w, "Failed to get friends", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"friends":  friends,
		})
	}
}
