package api

import (
	"encoding/json"
	"log"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateUser called")

	w.Header().Set(ContentType, ApplicationJSON)

	if r.Method != http.MethodPost {
		log.Printf("CreateUser - Method not allowed: %s", r.Method)
		errorResponse(w, http.StatusMethodNotAllowed, MethodErr)
		return
	}

	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		errorResponse(w, http.StatusBadRequest, JsonErr)
		return
	}

	if err := h.validateUser(newUser.Username, newUser.Password); err != nil {
		errorResponse(w, http.StatusConflict, err.Error())
		return
	}

	if err := h.createUser(&newUser); err != nil {
		log.Printf("Error creating user: %v", err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.storeUser(&newUser); err != nil {
		log.Printf("Error storing user: %v", err)
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
	successResponse(w, userResult)
	log.Printf("User %s created", newUser.ID)
}

func (h *Handler) RetrieveUserByID(w http.ResponseWriter, r *http.Request) {
	log.Println("RetrieveUserByID - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	if r.Method != http.MethodGet {
		log.Printf("RetrieveUserByID - Method not allowed: %s", r.Method)
		errorResponse(w, http.StatusMethodNotAllowed, MethodErr)
		return
	}

	userIDFromURL := r.URL.Path[len("/v2/users/"):]

	retrievedUser, err := h.getUserByID(userIDFromURL)
	if err != nil {
		log.Printf("RetrieveUserByID - User not found: %v", err)
		errorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	log.Printf("RetrieveUserByID - User %s retrieved successfully", userIDFromURL)

	successResponsePayload := models.SuccessRespons{
		Status: true,
		Result: retrievedUser,
	}

	successResponse(w, successResponsePayload)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("UpdateUser - called")
	w.Header().Set(ContentType, ApplicationJSON)

	if r.Method != http.MethodPut {
		log.Printf("UpdateUser - Method not allowed: %s", r.Method)
		errorResponse(w, http.StatusMethodNotAllowed, MethodErr)
		return
	}

	var userInfo models.User
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		log.Printf("UpdateUser - Invalid JSON format: %v", err)
		errorResponse(w, http.StatusBadRequest, JsonErr)
		return
	}

	updatedUser, err := h.update(&userInfo)
	if err != nil {
		log.Printf("UpdateUser - Internal Server Error: %v", err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("UpdateUser - User %s updated successfully", userInfo.ID)

	userResult := models.SuccessRespons{
		Status: true,
		Result: updatedUser,
	}
	successResponse(w, userResult)
}
