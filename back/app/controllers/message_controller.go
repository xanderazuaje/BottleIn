package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/xanderazuake/bottlenet/backend/app/config"
	"github.com/xanderazuake/bottlenet/backend/app/models"
	"go.mongodb.org/mongo-driver/bson"
)

// getRandomUser retrieves a random user from the database, optionally excluding a specific user
func getRandomUser(excludeID primitive.ObjectID) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.DB.Collection("users")

	// Build the filter to exclude the specified user
	filter := bson.M{}
	if !excludeID.IsZero() {
		filter["_id"] = bson.M{"$ne": excludeID}
	}

	// Get the count of documents matching the filter
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return models.User{}, err
	}

	if count == 0 {
		return models.User{}, fmt.Errorf("no users available")
	}

	// Generate a random skip number
	randomSkip := rand.Int63n(count)

	// Find one user after skipping randomSkip users
	opts := options.FindOne().SetSkip(randomSkip)

	var user models.User
	err = collection.FindOne(ctx, filter, opts).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// CreateMessage handles creating a new "bottle message"
// @Summary      Create a new bottle message
// @Description  Sends a bottle message to a random user
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        message  body      models.Message  true  "Message data"
// @Success      201  {object}   models.Message
// @Failure      400  {object}   string
// @Router       /messages/new [post]
func CreateMessage(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a Message struct
	var message models.Message
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("Error decoding message payload: %v\n", err)
		return
	}

	// Fetch the sender to ensure they exist
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sender models.User
	err = config.DB.Collection("users").FindOne(ctx, bson.M{"_id": message.SenderID}).Decode(&sender)
	if err != nil {
		http.Error(w, "Sender not found", http.StatusBadRequest)
		log.Printf("Error finding sender: %v\n", err)
		return
	}

	// Generate a random recipient (user), excluding the sender
	randomUser, err := getRandomUser(message.SenderID)
	if err != nil {
		http.Error(w, "Failed to assign a random user", http.StatusInternalServerError)
		log.Printf("Error generating random user: %v\n", err)
		return
	}

	// Set the recipient ID and timestamp
	message.RecipientID = randomUser.ID
	message.Timestamp = time.Now().Unix()

	// Insert the message into the database
	collection := config.DB.Collection("messages")
	result, err := collection.InsertOne(ctx, message)
	if err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		log.Printf("Error inserting message into DB: %v\n", err)
		return
	}

	// Return the created message with its ID
	message.ID = result.InsertedID.(primitive.ObjectID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(message)
	if err != nil {
		log.Printf("Error encoding message response: %v\n", err)
	}
}

// RespondToMessage allows the recipient to respond to the sender
// @Summary      Respond to a bottle message
// @Description  Responds to an existing bottle message
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Message ID"
// @Param        response  body      models.Message  true  "Response message"
// @Success      200  {object}   models.Message
// @Failure      400  {object}   string
// @Router       /messages/{id}/respond [post]
func RespondToMessage(w http.ResponseWriter, r *http.Request) {
	messageID, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		log.Printf("Invalid message ID: %v\n", err)
		return
	}

	// Parse the response message
	var responseMessage models.Message
	err = json.NewDecoder(r.Body).Decode(&responseMessage)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("Error decoding response message payload: %v\n", err)
		return
	}

	// Fetch the original message
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.DB.Collection("messages")
	var originalMessage models.Message
	err = collection.FindOne(ctx, bson.M{"_id": messageID}).Decode(&originalMessage)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		log.Printf("Error finding original message: %v\n", err)
		return
	}

	// Set up the response message
	responseMessage.SenderID = originalMessage.RecipientID
	responseMessage.RecipientID = originalMessage.SenderID
	responseMessage.Timestamp = time.Now().Unix()

	// Add the response to a thread
	err = addMessageToThread(originalMessage, responseMessage)
	if err != nil {
		http.Error(w, "Failed to add message to thread", http.StatusInternalServerError)
		log.Printf("Error adding message to thread: %v\n", err)
		return
	}

	// Return the response message
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(responseMessage)
	if err != nil {
		log.Printf("Error encoding response message: %v\n", err)
	}
}

// Add message to an existing or new thread
func addMessageToThread(original models.Message, response models.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	threads := config.DB.Collection("threads")

	// Check if a thread exists for this message
	var thread models.Thread
	if original.ThreadID != nil {
		err := threads.FindOne(ctx, bson.M{"_id": original.ThreadID}).Decode(&thread)
		if err != nil {
			return err
		}
	} else {
		// Create a new thread
		thread = models.Thread{
			ID:           primitive.NewObjectID(),
			Participants: []primitive.ObjectID{original.SenderID, original.RecipientID},
			Messages:     []primitive.ObjectID{original.ID},
		}

		// Insert the new thread into the database
		_, err := threads.InsertOne(ctx, thread)
		if err != nil {
			return err
		}

		// Update the original message to reference this thread
		messages := config.DB.Collection("messages")
		_, err = messages.UpdateOne(ctx, bson.M{"_id": original.ID}, bson.M{"$set": bson.M{"thread_id": thread.ID}})
		if err != nil {
			return err
		}
	}

	// Add the response message to the thread
	_, err := config.DB.Collection("messages").InsertOne(ctx, response)
	if err != nil {
		return err
	}

	// Update the thread with the new response message
	_, err = threads.UpdateOne(ctx, bson.M{"_id": thread.ID}, bson.M{"$push": bson.M{"messages": response.ID}})
	return err
}

// DropMessage allows the recipient to "drop" a message and send it to another random user
// @Summary      Drop the message into the sea
// @Description  Sends the message to another random user
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Message ID"
// @Success      200  {object}   models.Message
// @Failure      400  {object}   string
// @Router       /messages/{id}/drop [post]
func DropMessage(w http.ResponseWriter, r *http.Request) {
	messageID, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		log.Printf("Invalid message ID: %v\n", err)
		return
	}

	// Fetch the original message
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messages := config.DB.Collection("messages")
	var message models.Message
	err = messages.FindOne(ctx, bson.M{"_id": messageID}).Decode(&message)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		log.Printf("Error finding message: %v\n", err)
		return
	}

	// Generate a new random recipient
	randomUser, err := getRandomUser(message.SenderID)
	if err != nil {
		http.Error(w, "Failed to assign a new random user", http.StatusInternalServerError)
		log.Printf("Error generating random user: %v\n", err)
		return
	}

	// Update the recipient ID of the message
	_, err = messages.UpdateOne(ctx, bson.M{"_id": message.ID}, bson.M{"$set": bson.M{"recipient_id": randomUser.ID}})
	if err != nil {
		http.Error(w, "Failed to drop the message to a new user", http.StatusInternalServerError)
		log.Printf("Error updating message recipient: %v\n", err)
		return
	}

	// Return the updated message
	message.RecipientID = randomUser.ID
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(message)
	if err != nil {
		log.Printf("Error encoding dropped message: %v\n", err)
	}
}

// KeepMessage allows the user to keep the message in their account
// @Summary      Keep a message
// @Description  Saves the message in the user's account
// @Tags         messages
// @Param        id      path      string  true  "Message ID"
// @Param        userId  query     string  true  "User ID"
// @Success      200  {object}   string
// @Failure      400  {object}   string
// @Router       /messages/{id}/keep [get]
func KeepMessage(w http.ResponseWriter, r *http.Request) {
	messageID, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Retrieve the user ID from the query parameter
	userID, err := primitive.ObjectIDFromHex(r.URL.Query().Get("userId"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Add the message to the user's kept messages
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users := config.DB.Collection("users")
	_, err = users.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$addToSet": bson.M{"kept_messages": messageID}})
	if err != nil {
		http.Error(w, "Failed to keep the message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("Message successfully kept"))
}
