package api

import (
	"net/http"
	"strconv"

	models "github.com/Dzdrgl/redis-Api/Models"
)

func GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		ErrorRespons(w, 405, "Method not allowed")
		return
	}

	idStr := r.URL.Path[len("/v2/users/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrorRespons(w, 400, "Invalid ID")
		return
	}

	user, err := SearchByUserID(id)
	if err != nil {
		ErrorRespons(w, 404, err.Error())
		return
	}

	userMap := map[string]interface{}{
		"ID":       user.ID,
		"Username": user.Username,
		"Name":     user.Name,
		"Surname":  user.Surname,
	}

	userResult := models.SuccessRespons{
		Status: true,
		Result: userMap,
	}

	SuccessRespons(w, userResult)
}
