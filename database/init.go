package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitializeCollections sets up the MongoDB collections with proper validation
func InitializeCollections(ctx context.Context, db *mongo.Database) error {
	// Check if users collection exists
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": "users"})
	if err != nil {
		return err
	}

	// Create validator
	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"email", "username", "password", "first_name", "last_name", "created_at", "updated_at"},
			"properties": bson.M{
				"email": bson.M{
					"bsonType": "string",
					"pattern":  "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
				},
				"username": bson.M{
					"bsonType":  "string",
					"minLength": 3,
				},
				"password": bson.M{
					"bsonType":  "string",
					"minLength": 6,
				},
				"phone": bson.M{
					"bsonType": "string",
				},
				"code_id": bson.M{
					"bsonType": "string",
				},
				"first_name": bson.M{
					"bsonType": "string",
				},
				"last_name": bson.M{
					"bsonType": "string",
				},
				"birthdate": bson.M{
					"bsonType": "string",
					"pattern":  "^\\d{4}-\\d{2}-\\d{2}$",
				},
				"created_at": bson.M{
					"bsonType": "date",
				},
				"updated_at": bson.M{
					"bsonType": "date",
				},
				"deleted_at": bson.M{
					"bsonType": []string{"date", "null"},
				},
			},
		},
	}

	// If collection doesn't exist, create it
	if len(collections) == 0 {
		opts := options.CreateCollection().SetValidator(validator)
		if err := db.CreateCollection(ctx, "users", opts); err != nil {
			return err
		}
	} else {
		// If collection exists, update the validator
		cmd := bson.D{
			{Key: "collMod", Value: "users"},
			{Key: "validator", Value: validator},
			{Key: "validationLevel", Value: "strict"},
		}
		if err := db.RunCommand(ctx, cmd).Err(); err != nil {
			return err
		}
	}

	// Create or update indexes
	users := db.Collection("users")
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys:    bson.D{{Key: "code_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}

	// Drop existing indexes except _id
	if _, err := users.Indexes().DropAll(ctx); err != nil {
		return err
	}

	// Create new indexes
	_, err = users.Indexes().CreateMany(ctx, indexes)
	return err
}
