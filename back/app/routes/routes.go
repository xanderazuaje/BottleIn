// backend/routes/routes.go
package routes

import (
	"backend/controllers"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupRoutes initializes the router and routes
func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/hello", controllers.HelloHandler).Methods(http.MethodGet)

	return router
}
