package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	models "github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateUser called")

	w.Header().Set(ContentType, ApplicationJSON)

	if r.Method != http.MethodPost {
		log.Printf("UpdateUser - Method not allowed: %s", r.Method)
		errorResponse(w, http.StatusMethodNotAllowed, MethodNotAllowedMsg)
		return
	}

	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		log.Printf("UpdateUser - Invalid JSON format: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}

	if err := h.validateUser(newUser.Username, newUser.Password); err != nil {
		errorResponse(w, http.StatusConflict, err.Error())
		return
	}

	token, err := h.CreateUserInRedis(&newUser)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
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
	log.Printf("User %s created", newUser.ID)
}

func (h *Handler) HandleRetrieveUser(w http.ResponseWriter, r *http.Request) {
	log.Println("RetriveUser - called")
	w.Header().Set(ContentType, ApplicationJSON)

	if r.Method != http.MethodGet {
		log.Printf("RetriveUser - Method not allowed: %s", r.Method)
		errorResponse(w, http.StatusMethodNotAllowed, MethodNotAllowedMsg)
		return
	}

	userIDFromURL := r.URL.Path[len("/v1/users/"):]
	idToInt, err := strconv.Atoi(userIDFromURL)
	if err != nil {
		errorResponse(w, http.StatusNotFound, InvalidID)
		return
	}
	retrievedToken, err := h.GetTokenByID(idToInt)
	if err != nil {
		errorResponse(w, http.StatusNotFound, IDNotFound)
		return
	}
	user, err := h.FetchUserInfo(retrievedToken)
	if err != nil {
		errorResponse(w, http.StatusNotFound, err.Error())
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

	token := r.Header.Get("Authorization")
	if h.ValidateToken(token) == false {
		errorResponse(w, http.StatusNotFound, InvalidTokenMsg)
		return
	}

	if r.Method != http.MethodPut {
		log.Printf("UpdateUser - Method not allowed: %s", r.Method)
		errorResponse(w, http.StatusMethodNotAllowed, MethodNotAllowedMsg)
		return
	}

	var userInfo models.User
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		log.Printf("UpdateUser - Invalid JSON format: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}

	updatedUser, err := h.UpdateUser(&userInfo, token)
	if err != nil {
		log.Printf("UpdateUser - Internal Server Error: %v", err)
		errorResponse(w, http.StatusInternalServerError, InternalServerErrorMsg)
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
	if r.Method != http.MethodPost {
		log.Printf("UpdateUser - Method not allowed: %s", r.Method)
		errorResponse(w, http.StatusMethodNotAllowed, MethodNotAllowedMsg)
		return
	}
	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		log.Printf("UpdateUser - Invalid JSON format: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
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
