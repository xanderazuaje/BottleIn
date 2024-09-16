package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Message represents a "bottle message"
type Message struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	SenderID    primitive.ObjectID  `bson:"sender_id" json:"senderId"`
	RecipientID primitive.ObjectID  `bson:"recipient_id" json:"recipientId"`
	Content     string              `bson:"content" json:"content" example:"A message in a bottle"`
	Timestamp   int64               `bson:"timestamp" json:"timestamp"`
	ThreadID    *primitive.ObjectID `bson:"thread_id,omitempty" json:"threadId"`
}
