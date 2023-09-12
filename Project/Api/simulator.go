package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/Dzdrgl/redis-Api/models"
)

type SimInfo struct {
	Usercount int `json:"usercount"`
}

func (h *Handler) Simulator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		log.Printf("Method not allowed: %s", r.Method)
		return
	}

	var simInfo SimInfo
	if err := json.NewDecoder(r.Body).Decode(&simInfo); err != nil {
		errorResponse(w, http.StatusBadRequest, "JSON decode error")
		log.Printf("JSON decode error: %v", err)
		return
	}
	for i := 1; i <= simInfo.Usercount; i++ {
		if err := h.createSimUser(); err != nil {
			errorResponse(w, http.StatusInternalServerError, "Failed to create user")
			log.Printf("Failed to create user: %v", err)
			return
		}
	}
	log.Printf("%d user(s) created.", simInfo.Usercount)

	if err := h.matchSimulator(); err != nil {
		errorResponse(w, http.StatusInternalServerError, "Match operation failed")
		log.Printf("Match operation failed: %v", err)
		return
	}

	result := models.SuccessRespons{
		Status: true,
		Result: nil,
	}
	successResponse(w, result)
}

func (h *Handler) matchSimulator() error {
	userCount, err := h.client.Get("user_id").Result()
	if err != nil {
		return fmt.Errorf("User count error: %v", err)
	}

	countInt, err := strconv.Atoi(userCount)
	if err != nil {
		return fmt.Errorf("Conversion error: %v", err)
	}

	for i := 1; i < countInt; i++ {
		for j := i + 1; j <= countInt; j++ {
			var matchInfo MatchInfo
			matchInfo.FirstUserId = i
			matchInfo.FirstUserScore = rand.Intn(10)
			matchInfo.SecondUserId = j
			matchInfo.SecondUserScore = rand.Intn(10)

			if err := h.updateScore(matchInfo); err != nil {
				return fmt.Errorf("Failed to update score: %v", err)
			}
		}
	}

	log.Printf("All players got matched.")
	return nil
}

const defaultPassword = "123456"

func (h *Handler) createSimUser() error {
	newUserID := h.nextID()
	if newUserID == "" {
		return fmt.Errorf("UserID is empty")
	}
	username := fmt.Sprintf("player_%s", newUserID)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	name, surname := randomNameAndSurname()

	simUser := models.User{
		ID:       newUserID,
		Username: username,
		Password: string(hashedPass),
		Name:     name,
		Surname:  surname,
	}
	if err := h.storeUser(&simUser); err != nil {
		return err
	}
	return nil
}
