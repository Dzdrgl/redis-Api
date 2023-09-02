package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/Models"
)

func UserLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		ErrorRespons(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		ErrorRespons(w, http.StatusInternalServerError, "Error reading request")
		return
	}
	var creds models.User
	err = json.Unmarshal(body, &creds)
	if err != nil {
		ErrorRespons(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	user, err := Login(creds.Username, creds.Password)

	if err != nil {
		ErrorRespons(w, http.StatusUnauthorized, err.Error())
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

	SuccessRespons(w, userResult)
}
