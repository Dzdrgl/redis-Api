package api

import (
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

	models "github.com/Dzdrgl/redis-Api/models"
	"golang.org/x/crypto/bcrypt"
)

// ? CONSTANTS
const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"

	InvalidTokenMsg        = "Invalid token"
	InvalidJSONInputMsg    = "Invalid JSON input"
	MethodNotAllowedMsg    = "Method not allowed"
	InternalServerErrorMsg = "Internal server error"
	UsernameAlreadyExists  = "Username already exists"
	InvalidID              = "Invalid ID Format"
	IDNotFound             = "User Id not found"
)
const (
	MethodErr = "Method not allowed"
	JsonErr   = "Invalid JSON format"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Handler struct {
	client *redis.Client
}

func NewHandler(client *redis.Client) *Handler {
	return &Handler{client: client}
}

// !Create
func (h *Handler) CreateUserInRedis(user *models.User) (string, error) {
	if err := h.validateUser(user.Username, user.Password); err != nil {
		return "", err
	}
	user.ID = h.generateNextUserID()
	hashedPassword, err := HashPassword(user.Password)

	user.Password = hashedPassword
	userToken, err := h.storeUserInRedis(user)
	if err != nil {
		return "", err
	}
	return userToken, nil
}

func (h *Handler) storeUserInRedis(newUser *models.User) (string, error) {
	userToken := h.createToken()
	userKey := fmt.Sprintf("user:%s", userToken)
	keyId := fmt.Sprintf("userID:%s", newUser.ID)
	keyUsername := fmt.Sprintf("username:%s", newUser.Username)
	pipe := h.client.Pipeline()
	pipe.HMSet(userKey, map[string]interface{}{
		"id":       newUser.ID,
		"username": newUser.Username,
		"password": newUser.Password,
		"name":     newUser.Name,
		"surname":  newUser.Surname,
	})
	pipe.Set(keyUsername, userToken, 0)
	pipe.Set(keyId, userToken, 0)
	_, err := pipe.Exec()
	if err != nil {
		return "", err
	}
	return userToken, nil
}

func (h *Handler) validateUser(username, password string) error {
	if username == "" || password == "" {
		return errors.New("Username and password must not be empty")
	}
	key := fmt.Sprintf("username:%s", username)
	_, err := h.client.Get(key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Redis error: %s", err.Error())
	}
	return errors.New("Username already exists")
}

func (h *Handler) generateNextUserID() string {
	userId, err := h.client.Incr("user_id").Result()
	if err != nil {
		return ""
	}
	return strconv.FormatInt(userId, 10)
}

// !Fetch

func (h *Handler) FetchUserField(token, field string) (string, error) {
	key := fmt.Sprintf("user:%s", token)
	val, err := h.client.HGet(key, field).Result()
	if err != nil {
		return "", err
	}
	if val == "" {
		return "", fmt.Errorf("User not found")
	}

	return val, nil
}

func (h *Handler) FetchUserInfo(token string) (*models.User, error) {
	key := fmt.Sprintf("user:%s", token)
	user, err := h.client.HGetAll(key).Result()
	if err != nil {
		return nil, err
	}
	return mapToUser(user), nil
}
func mapToUser(val map[string]string) *models.User {
	return &models.User{
		ID:       val["id"],
		Username: val["username"],
		Name:     val["name"],
		Surname:  val["surname"],
	}
}

// ! Login
func (h *Handler) UserLogin(username, password string) (*string, error) {
	token, err := h.GetTokenByUsername(username)
	if err != nil {
		return nil, err
	}

	hashedPass, err := h.FetchUserField(token, "password")
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password)); err != nil {
		return nil, fmt.Errorf("Incorrect password")
	}

	return &token, nil
}

// !Update
func (h *Handler) UpdateUser(userInfo *models.User, token string) (*models.User, error) {

	userKey := fmt.Sprintf("user:%s", token)
	oldUsername, err := h.FetchUserField(token, "username")
	if err != nil {
		return nil, err
	}
	if userInfo.Username != "" {
		usernameKey := fmt.Sprintf("username:%s", userInfo.Username)
		_, err := h.client.Get(usernameKey).Result()
		if err == redis.Nil {
			h.client.Set(usernameKey, token, 0)
			h.client.HSet(userKey, "username", userInfo.Username)
			oldKey := fmt.Sprintf("username:%s", oldUsername)
			h.client.Del(oldKey)
		} else if err != nil {
			return nil, fmt.Errorf("Unknown error: %w", err)
		} else {
			return nil, fmt.Errorf(UsernameAlreadyExists)
		}
	}

	if userInfo.Password != "" {
		hashedPass, err := HashPassword(userInfo.Password)
		if err != nil {
			return nil, err
		}
		h.client.HSet(userKey, "password", hashedPass)
	}

	if userInfo.Name != "" {
		h.client.HSet(userKey, "name", userInfo.Name)
	}

	if userInfo.Surname != "" {
		h.client.HSet(userKey, "surname", userInfo.Surname)
	}
	updatedUser, err := h.FetchUserInfo(token)
	if err != nil {
		return nil, err
	}
	return updatedUser, nil
}

// ! SECTION: SIMULATION HANDLER FUNCTIONS
// ?Len keys user:* and
func (h *Handler) matchSimulation() error {

	val, err := h.client.Keys("user:*").Result()
	if err != nil {
		return err
	}
	userCount := len(val)
	log.Println(userCount)
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

	newUserID := h.generateNextUserID()
	fmt.Println(newUserID)
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
	_, err = h.storeUserInRedis(&simUser)
	if err != nil {
		return err
	}
	return nil
}

// !Match
func (h *Handler) UpdateScore(match models.MatchInfo) error {
	if match.FirstUserId == match.SecondUserId {
		return fmt.Errorf("User ID's are same")
	}
	firstUserToken, err := h.GetTokenByID(match.FirstUserId)
	if err != nil {
		return errors.New("First user does not exist")
	}
	secondUserToken, err := h.GetTokenByID(match.SecondUserId)
	if err != nil {
		return errors.New("Second user does not exist")
	}

	if match.FirstUserScore > match.SecondUserScore {
		if err := h.incScore(firstUserToken, 3); err != nil {
			log.Println("birinci kullanici ")
			return err
		}
	} else if match.FirstUserScore < match.SecondUserScore {
		if err := h.incScore(secondUserToken, 3); err != nil {
			return err
		}
	} else {
		err := h.incScore(firstUserToken, 1)
		err = h.incScore(secondUserToken, 1)
		if err != nil {
			return err
		}
	}
	return nil
}
func (h *Handler) incScore(token string, points int) error {
	_, err := h.client.ZIncrBy("leaderboard", float64(points), token).Result()
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

func (h *Handler) BuildLeaderboardList(leaderbordInfo models.LeaderbordInfo) ([]LeaderbordModel, error) {
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
	errorResponse := models.ErrorResponse{Status: false, Message: message}
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

// ! MIDDLEWARE : TOKEN AND VALIDATE
// ? base64, jwt, go kütüphanlerini araştır
func (h *Handler) createToken() string {
	var token string
	alfanumeric := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for {
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
		key := fmt.Sprintf("user:%s", token)
		_, err := h.client.Get(key).Result()
		if err == redis.Nil {
			break
		} else {
			continue
		}
	}
	return token
}
func (h *Handler) GetTokenByUsername(username string) (string, error) {
	key := fmt.Sprintf("username:%s", username)
	token, err := h.client.Get(key).Result()
	if err == redis.Nil || token == "" {
		return "", errors.New("Username not found")
	} else if err != nil {
		return "", err
	}
	return token, nil
}
func (h *Handler) GetTokenByID(id int) (string, error) {
	key := fmt.Sprintf("userID:%d", id)
	token, err := h.client.Get(key).Result()
	if err == redis.Nil || token == "" {
		return "", errors.New("User Id not found")
	} else if err != nil {
		return "", err
	}
	return token, nil
}
func (h *Handler) ValidateToken(token string) bool {
	key := fmt.Sprintf("user:%s", token)
	if token == "" {

		return false
	}
	isExist, _ := h.client.HExists(key, "id").Result()
	log.Println(isExist)
	if isExist == false {
		return false
	}
	return true
}

func HashPassword(password string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Error hashing the password: %v", err)
	}
	return string(hashedPass), nil
}

func AuthMidware(handleFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		key := fmt.Sprintf("user:%s", token)
		log.Println(key)

	}
}
