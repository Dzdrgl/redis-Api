package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/models"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) UserLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		errorResponse(w, http.StatusBadRequest, "Error unmarshaling JSON")
		return
	}

	user, err := h.login(creds.Username, creds.Password)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	userMap := map[string]interface{}{
		"ID":       user.ID,
		"Username": user.Username,
	}

	userResult := models.SuccessRespons{
		Status: true,
		Result: userMap,
	}

	successResponse(w, userResult)
}

func (h *Handler) login(username, password string) (*models.User, error) {

	userID, err := h.getIDByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving user ID")
	}

	key := fmt.Sprintf("user%s", userID)
	hashedPass, err := h.client.HGet(key, "password").Result()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password)); err != nil {
		return nil, fmt.Errorf("Incorrect password")
	}

	_, err = h.client.Set("currentUserId", userID, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("Error setting current user ID: %v", err)
	}

	user, err := h.getUser(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
