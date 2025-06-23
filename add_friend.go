package chat_package

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (adder *UserManager) AddFriend(c *Client, friend string) {
	adder.Mu.Lock()
	defer adder.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := adder.Client.Database("ChatDB").Collection("users")

	filter := map[string]interface{}{"username": c.Username}
	update := map[string]interface{}{
		"$addToSet": map[string]interface{}{
			"friends": friend,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error adding friend:", err)
		return
	}
}

func Add_friend(client *mongo.Client, adder *UserManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(w)
		username := r.URL.Query().Get("username")
		friend := r.URL.Query().Get("friend")
		if username == "" || friend == "" {
			http.Error(w, "username and friend are required", http.StatusBadRequest)
			return
		}
		clientObj := &Client{Username: username}
		adder.AddFriend(clientObj, friend)
		w.Write([]byte("Friend added successfully"))
	}
}
