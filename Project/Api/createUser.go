package api

import (
	"encoding/json"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/Models"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		ErrorRespons(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		ErrorRespons(w, http.StatusBadRequest, "Error unmarshaling JSON")
		return
	}

	if err := ValidateUser(&newUser); err != nil {
		ErrorRespons(w, http.StatusConflict, err.Error())
		return
	}

	if err := StoreUser(&newUser); err != nil {
		ErrorRespons(w, http.StatusInternalServerError, err.Error())
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

	SuccessRespons(w, userResult)
}
