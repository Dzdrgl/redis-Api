package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Dzdrgl/redis-Api/models"
	"github.com/go-redis/redis"
)

type MatchInfo struct {
	Firstuserid     int `json:"firstuserid"`
	Seconduserid    int `json:"seconduserid"`
	Firstuserscore  int `json:"firstuserscore"`
	Seconduserscore int `json:"seconduserscore"`
}

type LbInfo struct {
	Count int64 `json:"count"`
	Page  int64 `json:"page"`
}

func (h *Handler) Leaderboard(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Println("Method not allowed")
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var lbInfo LbInfo
	if err := json.NewDecoder(r.Body).Decode(&lbInfo); err != nil {
		log.Println("Error decoding JSON:", err)
		errorResponse(w, http.StatusNotFound, "Invalid JSON payload")
		return
	}
	startIndex := lbInfo.Count * (lbInfo.Page - 1)
	endIndex := startIndex + lbInfo.Count - 1

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
	successResponse(w, result)
}

func (h *Handler) GetMatchInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var match MatchInfo
	if err := json.NewDecoder(r.Body).Decode(&match); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	if match.Firstuserid == match.Seconduserid {
		errorResponse(w, http.StatusBadRequest, "Both user IDs are the same!")
		return
	}

	if err := h.updateScore(match); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(w, models.SuccessRespons{Status: true, Result: nil})

}
func (h *Handler) updateScore(match MatchInfo) error {

	firstUserKey := fmt.Sprintf("user%d", match.Firstuserid)
	secondUserKey := fmt.Sprintf("user%d", match.Seconduserid)

	if exists, err := h.client.Exists(firstUserKey).Result(); err != nil || exists == 0 {
		return errors.New("First user does not exist")
	}

	if exists, err := h.client.Exists(secondUserKey).Result(); err != nil || exists == 0 {
		return errors.New("Second user does not exist")
	}

	if match.Firstuserscore > match.Seconduserscore {
		if err := h.incScore(firstUserKey, 3); err != nil {
			return err
		}
	} else if match.Firstuserscore < match.Seconduserscore {
		if err := h.incScore(secondUserKey, 3); err != nil {
			return err
		}
	}

	if err := h.incScore(firstUserKey, 1); err != nil {
		return err
	}
	if err := h.incScore(secondUserKey, 1); err != nil {
		return err
	}

	return nil
}

func (h *Handler) incScore(userKey string, points int) error {
	score, err := h.client.HIncrBy(userKey, "score", int64(points)).Result()
	if err != nil {
		return err
	}
	_, err = h.client.ZAdd("leaderboard", redis.Z{
		Score:  float64(score),
		Member: userKey,
	}).Result()

	return err
}

func (h *Handler) leaderboardList(leaderboard []string) ([]map[string]interface{}, error) {
	var list []map[string]interface{}
	for i, user := range leaderboard {
		id, err := h.getUserInfo(user, "id")
		if err != nil {
			return nil, err
		}
		username, err := h.getUserInfo(user, "username")
		if err != nil {
			return nil, err
		}

		userInfo := map[string]interface{}{
			"rank":     i + 1,
			"id":       id,
			"username": username,
		}
		list = append(list, userInfo)
	}
	return list, nil
}
