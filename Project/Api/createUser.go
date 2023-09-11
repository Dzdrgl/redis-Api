package api

import (
	"encoding/json"
	"log"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateUser called")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		errorResponse(w, http.StatusBadRequest, "Error unmarshaling JSON")
		return
	}
	if err := h.validateUser(newUser.Username, newUser.Password); err != nil {
		errorResponse(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.storeUser(&newUser); err != nil {
		log.Println(err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	userMap := map[string]interface{}{
		"ID":       newUser.ID,
		"Username": newUser.Username,
	}

	userResult := models.SuccessRespons{
		Status: true,
		Result: userMap,
	}
	log.Println("user" + newUser.ID + " created")
	successResponse(w, userResult)
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		errorResponse(w, 405, "Method not allowed")
		return
	}

	idStr := r.URL.Path[len("/v2/users/"):]

	user, err := h.getUser(idStr)
	if err != nil {
		errorResponse(w, 404, err.Error())
		return
	}

	userResult := models.SuccessRespons{
		Status: true,
		Result: user,
	}

	successResponse(w, userResult)
}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "PUT" {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var userInfo models.User
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	updatedUser, err := h.update(&userInfo)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	userResult := models.SuccessRespons{
		Status: true,
		Result: updatedUser,
	}
	successResponse(w, userResult)
}
