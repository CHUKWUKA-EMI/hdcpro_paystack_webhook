package utils

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectDB connects to the MongoDB database
func ConnectDB(dbURL string) (*mongo.Client, error) {
	log.Println("Connecting to mongodb")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(dbURL).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	// Send a ping to confirm a successful connection
	if err := client.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MongoDB!")

	return client, nil
}
