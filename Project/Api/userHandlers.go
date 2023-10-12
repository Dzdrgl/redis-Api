package api

import (
	"encoding/json"
	"log"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateUser called")
	w.Header().Set(ContentType, ApplicationJSON)

	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		log.Printf("UpdateUser - Invalid JSON format: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}
	if newUser.Username == "" || newUser.Password == "" {
		errorResponse(w, http.StatusBadRequest, "Username and password must not be empty")
		return
	}
	if err := h.CreateUser(&newUser, ""); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := models.User{
		ID:       newUser.ID,
		Username: newUser.Username,
		Name:     newUser.Name,
		Surname:  newUser.Surname,
	}

	userResult := models.SuccessResponse{
		Status: true,
		Result: result,
	}
	successResponse(w, userResult)
}

func (h *Handler) HandleRetrieveUser(w http.ResponseWriter, r *http.Request) {
	log.Println("RetriveUser - called")
	w.Header().Set(ContentType, ApplicationJSON)

	_, ok := r.Context().Value("userInfo").(models.User)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User info not found in context")
		return
	}
	userIDFromURL := r.URL.Path[len("/api/v2/users/"):]
	user := h.FetchUserInfoWithID(userIDFromURL)
	if user.Username == "" {
		errorResponse(w, http.StatusNotFound, "User does not exist")
		return
	}
	response := models.SuccessResponse{
		Status: true,
		Result: user,
	}
	successResponse(w, response)
}

func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("UpdateUser - called")
	w.Header().Set(ContentType, ApplicationJSON)

	var newInfo models.User
	if err := json.NewDecoder(r.Body).Decode(&newInfo); err != nil {
		log.Printf("UpdateUser - Invalid JSON format: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}
	user, ok := r.Context().Value("userInfo").(models.User)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}
	updatedUser, err := h.UpdateUser(&newInfo, user.ID)
	if err != nil {
		log.Printf("UpdateUser - Internal Server Error: %v", err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	userResult := models.SuccessResponse{
		Status: true,
		Result: updatedUser,
	}
	successResponse(w, userResult)
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("UserLogin - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		log.Printf("UpdateUser - Invalid JSON format: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}
	if creds.Username == "" || creds.Password == "" {
		errorResponse(w, http.StatusBadRequest, "Username and password must not be empty")
		return
	}

	token, err := h.UserLogin(creds.Username, creds.Password)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	userMap := map[string]interface{}{
		"Token": token,
	}

	userResult := models.SuccessResponse{
		Status: true,
		Result: userMap,
	}
	successResponse(w, userResult)
}
