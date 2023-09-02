package utilities

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/Models"
	"golang.org/x/crypto/bcrypt"
)

func HashingPassword(password string) ([]byte, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Error hashing the password: %v", err)
	}
	return hashedPass, nil
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := models.ErrorRespons{Status: false, Message: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

func WriteSuccessResponse(w http.ResponseWriter, result models.SuccessRespons) {
	result.Status = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)
}
