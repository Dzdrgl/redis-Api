package api

import (
	"encoding/json"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/Models"
)

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "PUT" {
		ErrorRespons(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var userInfo models.User
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		ErrorRespons(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	updatedUser, err := Update(&userInfo)
	if err != nil {
		ErrorRespons(w, http.StatusInternalServerError, err.Error())
		return
	}

	userMap := map[string]interface{}{
		"ID":       updatedUser.ID,
		"Username": updatedUser.Username,
		"Name":     updatedUser.Name,
		"Surname":  updatedUser.Surname,
	}

	userResult := models.SuccessRespons{
		Status: true,
		Result: userMap,
	}

	SuccessRespons(w, userResult)

}
