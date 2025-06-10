package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitializeCollections sets up the MongoDB collections with proper validation
func InitializeCollectionsV2(ctx context.Context, db *mongo.Database) error {
	// Drop existing collection if it exists
	if err := db.Collection("users").Drop(ctx); err != nil {
		return err
	}

	// Create users collection with schema validation
	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"email", "phone", "password", "last_log", "created_at", "updated_at"},
			"properties": bson.M{
				"email": bson.M{
					"bsonType": "string",
					"pattern":  "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
				},
				"phone": bson.M{
					"bsonType": "string",
				},
				"password": bson.M{
					"bsonType": "string",
				},
				"last_log": bson.M{
					"bsonType": "string",
					"pattern":  "^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}$",
				},
				"created_at": bson.M{
					"bsonType": "date",
				},
				"updated_at": bson.M{
					"bsonType": "date",
				},
			},
		},
	}

	// Create collection with validation
	opts := options.CreateCollection().SetValidator(validator)
	if err := db.CreateCollection(ctx, "users", opts); err != nil {
		return err
	}

	// Create indexes
	users := db.Collection("users")
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	}

	_, err := users.Indexes().CreateMany(ctx, indexes)
	return err
}
