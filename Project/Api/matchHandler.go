package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Dzdrgl/redis-Api/models"
)

func (h *Handler) FetchLeaderboardPage(w http.ResponseWriter, r *http.Request) {
	log.Println("Fetchleaderboard - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	if r.Method != http.MethodPost {
		log.Println("Fetchleaderboard - Method not allowed")
		errorResponse(w, http.StatusMethodNotAllowed, MethodErr)
		return
	}

	var leaderbordInfo models.LeaderbordInfo
	if err := json.NewDecoder(r.Body).Decode(&leaderbordInfo); err != nil {
		log.Printf("Fetchleaderboard - Invalid JSON format")
		errorResponse(w, http.StatusNotFound, JsonErr)
		return
	}

	startIndex := leaderbordInfo.Count * (leaderbordInfo.Page - 1)
	endIndex := startIndex + leaderbordInfo.Count - 1

	leaderboard, err := h.client.ZRevRange("leaderboard", startIndex, endIndex).Result()
	if err != nil {
		log.Println("Error fetching leaderboard:", err)
		errorResponse(w, http.StatusNotFound, "Could not fetch leaderboard")
		return
	}

	list, err := h.leaderboardList(leaderboard)
	if err != nil {
		log.Println("Error building leaderboard list:", err)
		errorResponse(w, http.StatusNotFound, "Could not build leaderboard list")
		return
	}

	result := models.SuccessRespons{
		Status: true,
		Result: list,
	}
	log.Println("Leaderboard page %d fetch successfuly.", leaderbordInfo.Count)
	successResponse(w, result)
}

func (h *Handler) GetMatchInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("GetMatchInfo - Called")
	w.Header().Set(ContentType, ApplicationJSON)

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

	if err := h.updateScore(match); err != nil {
		log.Printf("GetMatchInfo - update score failed : %s", err)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("GetMatchInfo - User score save succesfuly complated.")
	successResponse(w, models.SuccessRespons{Status: true, Result: nil})
}
