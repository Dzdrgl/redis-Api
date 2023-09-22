package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	log.Println("Leaderboard - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	token := r.Header.Get("Authorization")
	if h.ValidateToken(token) == false {
		errorResponse(w, http.StatusNotFound, InvalidTokenMsg)
		return
	}
	if r.Method != http.MethodPost {
		log.Println("Fetchleaderboard - Method not allowed")
		errorResponse(w, http.StatusMethodNotAllowed, MethodNotAllowedMsg)
		return
	}

	var leaderbordInfo models.LeaderbordInfo
	if err := json.NewDecoder(r.Body).Decode(&leaderbordInfo); err != nil {
		log.Printf("Fetchleaderboard - Invalid JSON format")
		errorResponse(w, http.StatusNotFound, InvalidJSONInputMsg)
		return
	}

	if leaderbordInfo.Count == 0 || leaderbordInfo.Page == 0 {
		log.Printf("Fetchleaderboard - The number of pages and users should not be zero..")
		errorResponse(w, http.StatusNotFound, "The number of pages and users should not be zero.")
		return
	}

	leaderboard, err := h.BuildLeaderboardList(leaderbordInfo)
	if err != nil {
		log.Println("Error building leaderboard list:", err)
		errorResponse(w, http.StatusNotFound, "Could not build leaderboard list")
		return
	}

	result := models.SuccessResponse{
		Status: true,
		Result: leaderboard,
	}
	log.Println("Leaderboard page %d fetch successfuly.", leaderbordInfo.Count)
	successResponse(w, result)
}

func (h *Handler) HandleMatch(w http.ResponseWriter, r *http.Request) {
	log.Println("GetMatchInfo - Called")
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

	var match models.MatchInfo
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		log.Printf("GetMatchInfo - Invalid JSON input")
		errorResponse(w, http.StatusBadRequest, JsonErr)
		return
	}

	if err := h.UpdateScore(match); err != nil {
		log.Printf("GetMatchInfo - update score failed : %s", err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("GetMatchInfo - User score save succesfuly complated.")
	successResponse(w, models.SuccessResponse{Status: true, Result: "User score save succesfuly"})
}
