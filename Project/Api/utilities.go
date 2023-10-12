package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"golang.org/x/crypto/bcrypt"

	models "github.com/Dzdrgl/redis-Api/models"
)

// ? CONSTANTS
const (
	ContentType            = "Content-Type"
	ApplicationJSON        = "application/json"
	InvalidTokenMsg        = "Invalid token"
	InvalidJSONInputMsg    = "Invalid JSON input"
	MethodNotAllowedMsg    = "Method not allowed"
	InternalServerErrorMsg = "Internal server error"
	UsernameAlreadyExists  = "Username already exists"
	InvalidID              = "Invalid ID Format"
	IDNotFound             = "User Id not found"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type IHandler interface {
	StoreUser(newUser *models.User) error
	CreateUser(newUser *models.User) error
	UserLogin(username, password string) (string, error)
}

type Handler struct {
	client *redis.Client
}

func NewHandler(redisClient *redis.Client) *Handler {
	return &Handler{
		client: redisClient,
	}
}

// !CreateUser
func (h *Handler) StoreUser(newUser *models.User) error {
	userKey := fmt.Sprintf("user:%s", newUser.ID)
	usernameKey := fmt.Sprintf("username:%s", newUser.Username)

	h.client.HMSet(userKey, map[string]interface{}{
		"id":       newUser.ID,
		"username": newUser.Username,
		"password": newUser.Password,
		"name":     newUser.Name,
		"surname":  newUser.Surname,
	}).Result()
	h.client.Set(usernameKey, newUser.ID, 0)
	return nil
}
func (h *Handler) CreateUser(newUser *models.User, id string) error {
	userID := h.GetUserIDWithUsername(newUser.Username)
	if userID != "" {
		return errors.New("Username already exsist.")
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Error hashing the password: %v", err)
	}
	newUser.Password = string(hashedPass)
	if id == "" {
		int64ID, _ := h.client.Incr("user_id").Result()
		intID := strconv.FormatInt(int64ID, 10)
		newUser.ID = string(intID)
	} else {
		newUser.ID = id
	}
	err = h.StoreUser(newUser)
	if err != nil {
		return err
	}
	return nil
}

// !!!!!!!!!!!!!!!!!!!!!!!!!<-Login->!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func (h *Handler) UserLogin(username, password string) (string, error) {
	userId := h.GetUserIDWithUsername(username)
	if userId == "" {
		return "", errors.New("User not found")
	}

	hashedPass := h.FetchUserFieldWithID(userId, "password")
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password)); err != nil {
		return "", fmt.Errorf("Incorrect password")
	}
	token := h.FetchUserFieldWithID(userId, "token")
	if token != "" {
		return token, nil
	}

	newToken, err := h.CreateToken(userId)
	if err != nil {
		return "", err
	}
	return newToken, nil
}

// !!!!!!!!!!!!!!!!!<--FetchUser-->!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func (h *Handler) FetchUserFieldWithID(id, field string) string {
	val := h.client.HGet("user:"+id, field).Val()
	return val
}
func (h *Handler) FetchUserInfoWithID(id string) *models.User {
	user := h.client.HGetAll("user:" + id).Val()
	return mapToUser(user)
}

func (h *Handler) GetUserIDWithUsername(username string) string {
	val := h.client.Get("username:" + username).Val()
	return val
}
func (h *Handler) FetchUserInfoWithToken(token string) (*models.User, error) {
	userID := h.client.Get("token:" + token).Val()
	if userID == "" {
		return nil, errors.New("Token does not found")
	}
	val := h.FetchUserFieldWithID(userID, "token")
	if val != token {
		return nil, errors.New("Invalid token")
	}
	user := h.FetchUserInfoWithID(userID)
	return user, nil
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!<-Update->!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func (h *Handler) UpdateUser(newInfo *models.User, id string) (*models.User, error) {
	oldName := h.FetchUserFieldWithID(id, "username")

	if newInfo.Username != "" {
		userID := h.GetUserIDWithUsername(newInfo.Username)
		if userID != "" {
			return nil, errors.New("Username already exist")
		} else if newInfo.Username == oldName {
			return nil, errors.New("Username already exist")
		}
		h.UpdateUserField(id, "username", newInfo.Username)
		h.client.Set("username:"+newInfo.Username, userID, 0)
		h.client.Del("username:" + oldName)

	}
	if newInfo.Password != "" {
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(newInfo.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("Error hashing the password: %v", err)
		}
		h.UpdateUserField(id, "password", string(hashedPass))
	}
	if newInfo.Name != "" {
		h.UpdateUserField(id, "name", newInfo.Name)
	}
	if newInfo.Surname != "" {
		h.UpdateUserField(id, "surname", newInfo.Surname)
	}

	return h.FetchUserInfoWithID(id), nil
}
func (h *Handler) UpdateUserField(id, field, newValue string) error {
	return h.client.HSet("user:"+id, field, newValue).Err()
}

// !!!!!!!!!!!!!!!!!!!!!!!!!<-Token->!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func randomToken() string {
	var token string
	alfanumeric := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	for i := 1; i <= 4; i++ {
		for j := 1; j <= 10; j++ {
			randIndex := rand.Intn(len(alfanumeric) - 1)
			randChar := string(alfanumeric[randIndex])
			token = token + randChar
		}
		if i != 4 {
			token = token + "-"
		}
	}
	return token
}

func (h *Handler) TokenIsExist(token string) bool {
	val, err := h.client.Get("token:" + token).Result()
	if err != nil {
		return false
	} else if val == "" {
		return false
	}
	return true
}
func (h *Handler) CreateToken(userID string) (string, error) {
	token := randomToken()
	isExsist := h.TokenIsExist(token)
	if isExsist == true {
		return h.CreateToken(userID)
	}
	err := h.UpdateUserField(userID, "token", token)
	err = h.client.Set("token:"+token, userID, 0).Err()
	if err != nil {
		return "", err
	}
	return token, nil
}

func mapToUser(val map[string]string) *models.User {
	return &models.User{
		ID:       val["id"],
		Username: val["username"],
		Name:     val["name"],
		Surname:  val["surname"],
	}
}

// ! SECTION: SIMULATION HANDLER FUNCTIONS
func (h *Handler) matchSimulation() error {
	userCount, err := h.client.Get("user_id").Int()
	if err != nil {
		return err
	}
	for i := 1; i < userCount; i++ {
		for j := i + 1; j <= userCount; j++ {
			var matchInfo models.MatchInfo
			matchInfo.FirstUserId = i
			matchInfo.FirstUserScore = rand.Intn(10)
			matchInfo.SecondUserId = j
			matchInfo.SecondUserScore = rand.Intn(10)

			if err := h.UpdateScore(matchInfo); err != nil {
				return fmt.Errorf("Failed to update score: %v", err)
			}
		}
	}

	log.Printf("All players got matched.")
	return nil

	//! Eger yeni kullanıcı eklendiğinde sadece eklenenler arasında maç yaptırmak için.
	/*
		!func (h *Handler) matchSimulation(userCount int) error {
		....
			for i := lastID ; i < userCount; i++ {
				for j := i + 1; j <= userCount; j++ {
					var matchInfo models.MatchInfo
					matchInfo.FirstUserId = i
					matchInfo.FirstUserScore = rand.Intn(10)
					matchInfo.SecondUserId = j
					matchInfo.SecondUserScore = rand.Intn(10)

					if err := h.updateScore(matchInfo); err != nil {
						return fmt.Errorf("Failed to update score: %v", err)
					}
				}
			}
		....
		}
	*/

}

func (h *Handler) createSimUser() error {
	const defaultPassword = "123456"
	int64ID, _ := h.client.Incr("user_id").Result()
	intID := strconv.FormatInt(int64ID, 10)
	stringID := string(intID)
	username := fmt.Sprintf("player_%s", stringID)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	name, surname := randomNameAndSurname()

	simUser := models.User{
		ID:       stringID,
		Username: username,
		Password: string(hashedPass),
		Name:     name,
		Surname:  surname,
	}
	if err = h.StoreUser(&simUser); err != nil {
		return err
	}
	return nil
}

// !Match
func (h *Handler) UpdateScore(match models.MatchInfo) error {
	if match.FirstUserId == match.SecondUserId {
		return fmt.Errorf("User ID's are same")
	}
	firstIdToStr := strconv.Itoa(match.FirstUserId)
	secondDdToStr := strconv.Itoa(match.SecondUserId)
	firstId := h.FetchUserFieldWithID(firstIdToStr, "id")
	if firstId == "" {
		return errors.New("First user does not exist")
	}
	secondId := h.FetchUserFieldWithID(secondDdToStr, "id")
	if secondId == "" {
		return errors.New("Second user does not exist")
	}

	if match.FirstUserScore > match.SecondUserScore {
		if err := h.incScore(firstId, 3); err != nil {
			log.Println("birinci kullanici ")
			return err
		}
	} else if match.FirstUserScore < match.SecondUserScore {

		if err := h.incScore(secondId, 3); err != nil {
			return err
		}
	} else {
		err := h.incScore(firstId, 1)
		err = h.incScore(secondId, 1)
		if err != nil {
			return err
		}
	}
	return nil
}
func (h *Handler) incScore(id string, points int) error {
	_, err := h.client.ZIncrBy("leaderboard", float64(points), id).Result()
	if err != nil {
		return err
	}
	return nil
}

// ! Leaderboard
type LeaderbordModel struct {
	Rank     int     `json:"rank"`
	Id       string  `json:"id"`
	Username string  `json:"username"`
	Score    float64 `json:"score"`
}

func (h *Handler) BuildLeaderboardList(leaderbordInfo models.ListInfo) ([]LeaderbordModel, error) {
	var leaderboard []LeaderbordModel
	startIndex := leaderbordInfo.Count * (leaderbordInfo.Page - 1)
	endIndex := startIndex + leaderbordInfo.Count - 1
	results, err := h.client.ZRevRangeWithScores("leaderboard", startIndex, endIndex).Result()
	if err != nil {
		return nil, err
	}
	for rank, user := range results {
		key := fmt.Sprintf("user:%s", user.Member.(string))
		var userInfo LeaderbordModel
		fields, err := h.client.HMGet(key, "id", "username").Result()
		if err != nil {
			return nil, err
		}
		userInfo.Id = fields[0].(string)
		userInfo.Username = fields[1].(string)
		userInfo.Rank = int(startIndex) + rank + 1
		userInfo.Score = user.Score

		leaderboard = append(leaderboard, userInfo)
	}

	return leaderboard, nil
}

// ! RANDOM NAME AND SURNAME CREATE.
type NameSurname struct {
	Names    []string `json:"names"`
	Surnames []string `json:"surnames"`
}

func randomNameAndSurname() (string, string) {
	file, err := ioutil.ReadFile("../namesAndSurnames.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return "", ""
	}

	var data NameSurname
	err = json.Unmarshal(file, &data)
	if err != nil {
		fmt.Println("Error unmarshalling:", err)
		return "", ""
	}

	randomName := data.Names[rand.Intn(len(data.Names))]
	randomSurname := data.Surnames[rand.Intn(len(data.Surnames))]

	return randomName, randomSurname
}

// ! ERROR AND SUCCES RESPONS
func errorResponse(w http.ResponseWriter, statusCode int, message string) {
	err := models.ErrorResponse{ErrorMessage: message}
	errorResponse := models.SuccessResponse{Status: false, Result: err}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}
func successResponse(w http.ResponseWriter, result models.SuccessResponse) {
	result.Status = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		user, err := h.FetchUserInfoWithToken(token)
		if err != nil || user == nil {
			errorResponse(w, http.StatusUnauthorized, "Authentication failed")
			return
		}
		userInfo := models.User{
			ID:       user.ID,
			Username: user.Username,
		}
		ctx := context.WithValue(r.Context(), "userInfo", userInfo)
		next(w, r.WithContext(ctx))
	}
}

//!!!!!!!!!!Friends

func (h *Handler) SentRequest(currentUser models.User, id string) error {
	now := time.Now().Unix()
	val, err := h.client.ZAdd("requests:"+id, redis.Z{
		Member: currentUser.ID,
		Score:  float64(now),
	}).Result()
	if err != nil {
		return errors.New("Error sending friend request")
	} else if val == 0 {
		return errors.New("Already sent friend request")
	}
	return nil
}
