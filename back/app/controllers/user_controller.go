package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/xanderazuake/bottlenet/backend/app/config"
	"github.com/xanderazuake/bottlenet/backend/app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateUser handles the creation of a new user
// @Summary      Create User
// @Description  Creates a new user in the database
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User data"
// @Success      201  {object}   models.User
// @Failure      400  {object}   string
// @Router       /users [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a User struct
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Set up a context for the MongoDB operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Insert the new user into MongoDB
	collection := config.DB.Collection("users")
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Error inserting user into DB", http.StatusInternalServerError)
		return
	}

	// Return the created user with the ID
	user.ID = result.InsertedID.(primitive.ObjectID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUsers handles retrieving all users from the database
// @Summary      Get Users
// @Description  Retrieves all users from the database
// @Tags         users
// @Produce      json
// @Success      200  {array}   models.User
// @Failure      500  {object}  string
// @Router       /users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Set up a context for the MongoDB operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Retrieve all users from MongoDB
	collection := config.DB.Collection("users")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Error fetching users from DB", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Decode the users into a slice
	var users []models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			http.Error(w, "Error decoding user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
