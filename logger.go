package pkg

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Logger struct {
	Mu     sync.Mutex
	Client *mongo.Client
}

var MessageLog = &Logger{}

var UserLog = &Logger{}

func (ml *Logger) LogMessage(msg ChatMessage) error {
	ml.Mu.Lock()
	defer ml.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := ml.Client.
		Database("ChatDB").
		Collection("messages")

	_, err := collection.InsertOne(ctx, msg)
	return err
}
