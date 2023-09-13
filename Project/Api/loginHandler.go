package api

import (
	"encoding/json"
	"log"
	"net/http"

	models "github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) UserLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("UserLogin - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	if r.Method != http.MethodPost {
		log.Println("UserLogin - Method Not Allowed")
		errorResponse(w, http.StatusMethodNotAllowed, MethodErr)
		return
	}
	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		log.Println("UserLogin - Invalid Json Input")
		errorResponse(w, http.StatusBadRequest, JsonErr)
		return
	}

	user, err := h.login(creds.Username, creds.Password)
	if err != nil {
		log.Println("UserLogin - Login failed :%s", err)
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
	log.Println("UserLogin - Login proccess complated.")
	successResponse(w, userResult)
}
