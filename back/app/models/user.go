package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User represents a user in the database
type User struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name         string               `bson:"name" json:"name" example:"RandomUser123"`
	Email        string               `bson:"email" json:"email" example:"user@example.com"`
	KeptMessages []primitive.ObjectID `bson:"kept_messages,omitempty" json:"keptMessages"`
}
