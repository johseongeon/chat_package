package chat_package

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserManager struct {
	Mu     sync.Mutex
	Client *mongo.Client
}

func RegisterUser(client *mongo.Client, username string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("ChatDB").Collection("users")

	filter := map[string]interface{}{"username": username}
	update := map[string]interface{}{
		"$setOnInsert": map[string]interface{}{
			"username": username,
			"friends":  []string{},
			"rooms":    []string{},
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error registering user:", err)
		return
	}

	fmt.Println("User registered:", username)
}

func RegisterServer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		RegisterUser(client, username)
		w.Write([]byte("User registered successfully"))
	}
}
