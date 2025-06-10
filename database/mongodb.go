package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

type Config struct {
	URI      string
	Database string
}

func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create client
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping database to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Printf("Connected to MongoDB at %s", uri)

	// Get database
	database := client.Database(dbName)

	// Initialize collections
	if err := InitializeCollectionsV2(ctx, database); err != nil {
		return nil, err
	}

	return &MongoDB{
		client:   client,
		database: database,
	}, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// Users returns the users collection
func (m *MongoDB) Users() *mongo.Collection {
	return m.database.Collection("users")
}
