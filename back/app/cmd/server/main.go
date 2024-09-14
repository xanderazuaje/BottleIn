// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/xanderazuake/bottlenet/backend/app/config"
	"github.com/xanderazuake/bottlenet/backend/app/routes"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/xanderazuake/bottlenet/backend/app/docs"
)

// @title           Go Backend API
// @version         1.0
// @description     This is a sample Go backend server.
// @host      localhost:8080
// @BasePath  /api

func main() {
	// Initialize MongoDB
	config.InitializeDB()

	// Initialize routes
	router := routes.SetupRoutes()

	// Swagger endpoint
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
