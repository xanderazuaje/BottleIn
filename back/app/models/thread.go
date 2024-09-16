package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Thread represents a message thread between two users
type Thread struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Messages     []primitive.ObjectID `bson:"messages" json:"messages"`         // Array of message IDs
	Participants []primitive.ObjectID `bson:"participants" json:"participants"` // Sender and recipient IDs
}
