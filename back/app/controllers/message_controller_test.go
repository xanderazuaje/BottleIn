// backend/controllers/message_controller_test.go
package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xanderazuake/bottlenet/backend/app/config"
	"github.com/xanderazuake/bottlenet/backend/app/models"
	"github.com/xanderazuake/bottlenet/backend/app/routes"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // Import this for ObjectIDs
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// createRouter initializes the router with all routes
func createRouter() *mux.Router {
	return routes.SetupRoutes()
}

// setupTestDB connects to the test MongoDB and inserts test users
func setupTestDB(t *testing.T) {
	// Set up a local MongoDB instance for testing
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatal("Failed to connect to MongoDB:", err)
	}

	// Assign the test database
	config.DB = client.Database("test_bottlenet")

	// Insert test users
	users := []interface{}{
		models.User{
			ID:           primitive.NewObjectID(),
			Name:         "TestUser1",
			Email:        "test1@example.com",
			KeptMessages: []primitive.ObjectID{},
		},
		models.User{
			ID:           primitive.NewObjectID(),
			Name:         "TestUser2",
			Email:        "test2@example.com",
			KeptMessages: []primitive.ObjectID{},
		},
	}

	_, err = config.DB.Collection("users").InsertMany(ctx, users)
	if err != nil {
		t.Fatal("Failed to insert test users:", err)
	}
}

// teardownTestDB cleans up the test database after tests
func teardownTestDB(t *testing.T) {
	// Clean up the test database after the tests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := config.DB.Drop(ctx)
	if err != nil {
		t.Fatal("Failed to drop test database:", err)
	}
}

func TestCreateMessage(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Fetch a test user to act as the sender
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sender models.User
	err := config.DB.Collection("users").FindOne(ctx, bson.M{"email": "test1@example.com"}).Decode(&sender)
	if err != nil {
		t.Fatal("Failed to fetch test sender:", err)
	}

	// Mock a request to create a new message
	body := []byte(`{"senderId":"` + sender.ID.Hex() + `","content":"Message in a bottle"}`)
	req, err := http.NewRequest("POST", "/api/messages/new", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal("Failed to create request:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set up the router
	router := createRouter()

	// Set up a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Execute the request via the router
	router.ServeHTTP(rr, req)

	// Check if the status code is 201 (Created)
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check if the response body contains the message
	var message models.Message
	err = json.NewDecoder(rr.Body).Decode(&message)
	if err != nil {
		t.Fatal("Failed to decode response:", err)
	}
	if message.Content != "Message in a bottle" {
		t.Errorf("Expected message content 'Message in a bottle', got %s", message.Content)
	}
	if message.SenderID != sender.ID {
		t.Errorf("Expected sender ID %s, got %s", sender.ID.Hex(), message.SenderID.Hex())
	}
	if message.RecipientID.IsZero() {
		t.Errorf("Expected recipient ID to be set, got zero value")
	}
}

func TestRespondToMessage(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Fetch test users to act as sender and recipient
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sender models.User
	err := config.DB.Collection("users").FindOne(ctx, bson.M{"email": "test1@example.com"}).Decode(&sender)
	if err != nil {
		t.Fatal("Failed to fetch test sender:", err)
	}

	var recipient models.User
	err = config.DB.Collection("users").FindOne(ctx, bson.M{"email": "test2@example.com"}).Decode(&recipient)
	if err != nil {
		t.Fatal("Failed to fetch test recipient:", err)
	}

	// Insert a mock message
	message := models.Message{
		ID:          primitive.NewObjectID(),
		SenderID:    sender.ID,
		Content:     "Initial message",
		Timestamp:   time.Now().Unix(),
		RecipientID: recipient.ID,
	}
	_, err = config.DB.Collection("messages").InsertOne(context.TODO(), message)
	if err != nil {
		t.Fatal("Failed to insert mock message:", err)
	}

	// Mock a request to respond to the message
	body := []byte(`{"content":"Response to the message"}`)
	req, err := http.NewRequest("POST", "/api/messages/"+message.ID.Hex()+"/respond", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal("Failed to create request:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set up the router
	router := createRouter()

	// Set up a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Execute the request via the router
	router.ServeHTTP(rr, req)

	// Check if the status code is 200 (OK)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check if the response body contains the response message
	var responseMessage models.Message
	err = json.NewDecoder(rr.Body).Decode(&responseMessage)
	if err != nil {
		t.Fatal("Failed to decode response:", err)
	}
	if responseMessage.Content != "Response to the message" {
		t.Errorf("Expected response message 'Response to the message', got %s", responseMessage.Content)
	}
	if responseMessage.SenderID != recipient.ID {
		t.Errorf("Expected sender ID %s, got %s", recipient.ID.Hex(), responseMessage.SenderID.Hex())
	}
	if responseMessage.RecipientID != sender.ID {
		t.Errorf("Expected recipient ID %s, got %s", sender.ID.Hex(), responseMessage.RecipientID.Hex())
	}
	if responseMessage.ThreadID.IsZero() {
		t.Errorf("Expected thread ID to be set, got zero value")
	}
}

func TestDropMessage(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Fetch test users to act as sender and recipient
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sender models.User
	err := config.DB.Collection("users").FindOne(ctx, bson.M{"email": "test1@example.com"}).Decode(&sender)
	if err != nil {
		t.Fatal("Failed to fetch test sender:", err)
	}

	var recipient models.User
	err = config.DB.Collection("users").FindOne(ctx, bson.M{"email": "test2@example.com"}).Decode(&recipient)
	if err != nil {
		t.Fatal("Failed to fetch test recipient:", err)
	}

	// Insert a mock message
	message := models.Message{
		ID:          primitive.NewObjectID(),
		SenderID:    sender.ID,
		Content:     "Message to be dropped",
		Timestamp:   time.Now().Unix(),
		RecipientID: recipient.ID,
	}
	_, err = config.DB.Collection("messages").InsertOne(context.TODO(), message)
	if err != nil {
		t.Fatal("Failed to insert mock message:", err)
	}

	// Mock a request to drop the message
	req, err := http.NewRequest("POST", "/api/messages/"+message.ID.Hex()+"/drop", nil)
	if err != nil {
		t.Fatal("Failed to create request:", err)
	}

	// Set up the router
	router := createRouter()

	// Set up a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Execute the request via the router
	router.ServeHTTP(rr, req)

	// Check if the status code is 200 (OK)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check if the message was dropped to a new recipient
	var updatedMessage models.Message
	err = config.DB.Collection("messages").FindOne(context.TODO(), bson.M{"_id": message.ID}).Decode(&updatedMessage)
	if err != nil {
		t.Fatal("Failed to fetch updated message:", err)
	}
	if updatedMessage.RecipientID == message.RecipientID {
		t.Errorf("Expected message recipient to change, but it remained the same")
	}
}
