// backend/routes/routes.go
package routes

import (
	"github.com/gorilla/mux"
	"github.com/xanderazuake/bottlenet/backend/app/controllers"
	"net/http"
)

// SetupRoutes initializes the router and defines the routes
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/hello", controllers.HelloHandler).Methods(http.MethodGet)

	// User routes
	api.HandleFunc("/users", controllers.CreateUser).Methods(http.MethodPost)
	api.HandleFunc("/users", controllers.GetUsers).Methods(http.MethodGet)

	// Message routes
	api.HandleFunc("/messages/new", controllers.CreateMessage).Methods(http.MethodPost)
	api.HandleFunc("/messages/{id}/respond", controllers.RespondToMessage).Methods(http.MethodPost)
	api.HandleFunc("/messages/{id}/drop", controllers.DropMessage).Methods(http.MethodPost)
	api.HandleFunc("/messages/{id}/keep", controllers.KeepMessage).Methods(http.MethodGet)

	return router
}
