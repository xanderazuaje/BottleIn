// backend/controllers/hello_controller.go
package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/xanderazuake/bottlenet/backend/app/models"
)

// HelloHandler handles the /api/hello endpoint
// @Summary      Say Hello
// @Description  Returns a hello message
// @Tags         hello
// @Produce      json
// @Success      200  {object}  models.Response
// @Router       /hello [get]
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	response := models.Response{Message: "Hello from Go Backend!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
