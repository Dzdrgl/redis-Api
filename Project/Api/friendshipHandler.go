package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) HandlerSearchUser(w http.ResponseWriter, r *http.Request) {
	log.Println("SearchUser - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	token := r.Header.Get("Authorization")
	if h.ValidateToken(token) == false {
		errorResponse(w, http.StatusNotFound, InvalidTokenMsg)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("GetMatchInfo - Method not allowed: %s", r.Method)
		errorResponse(w, http.StatusMethodNotAllowed, MethodErr)
		return
	}

	var userInfo *models.User
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		log.Printf("GetMatchInfo - Invalid JSON input")
		errorResponse(w, http.StatusBadRequest, JsonErr)
		return
	}
	val, err := h.FetchUserField(token, "username")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Println(val)
	log.Println(userInfo.Username)
	if val == userInfo.Username {
		errorResponse(w, http.StatusBadRequest, "Cant search your own name")
		return
	}

	token1, err := h.GetTokenByUsername(userInfo.Username)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	userId, err := h.FetchUserField(token1, "id")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	userMap := map[string]interface{}{
		"id": userId,
	}
	log.Printf("GetMatchInfo - User score save succesfuly complated.")
	successResponse(w, models.SuccessResponse{Status: true, Result: userMap})
}
