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

	"github.com/go-redis/redis"

	models "github.com/Dzdrgl/redis-Api/models"
	"golang.org/x/crypto/bcrypt"
)

// ? CONSTANTS
const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
	MethodErr       = "Method not allowed"
	JsonErr         = "Invalid JSON format"
)

type Handler struct {
	client *redis.Client
}

func NewHandler(client *redis.Client) *Handler {
	return &Handler{client: client}
}

// !SECTION: CREATE USER FUNCTIONS.
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

func (h *Handler) nextID() string {
	userId, err := h.client.Incr("user_id").Result()
	if err != nil {
		return ""
	}
	return strconv.FormatInt(userId, 10)
}

func hashingPassword(password string) ([]byte, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Error hashing the password: %v", err)
	}
	return hashedPass, nil
}

func (h *Handler) createUser(newUser *models.User) error {
	newUser.ID = h.nextID()
	if newUser.ID == "" {
		return fmt.Errorf("Failed to generate user ID")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	newUser.Password = string(hashedPass)

	return nil
}

func (h *Handler) storeUser(newUser *models.User) error {

	keyID := fmt.Sprintf("user%s", newUser.ID)
	keyUsername := fmt.Sprintf("username:%s", newUser.Username)

	pipe := h.client.Pipeline()
	pipe.HMSet(keyID, map[string]interface{}{
		"id":       newUser.ID,
		"username": newUser.Username,
		"password": newUser.Password,
		"name":     newUser.Name,
		"surname":  newUser.Surname,
		"score":    "0",
	})
	pipe.Set(keyUsername, newUser.ID, 0)
	_, err := pipe.Exec()
	if err != nil {
		return err
	}
	return nil
}

//!SECTION: SERACH AND GET USER

func (h *Handler) getIDByUsername(username string) (string, error) {
	key := fmt.Sprintf("username:%s", username)
	val, err := h.client.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (h *Handler) getUserByID(id string) (*models.User, error) {
	key := fmt.Sprintf("user%s", id)
	val, err := h.client.HGetAll(key).Result()
	if err != nil {
		log.Fatal(err)
	}

	if len(val) == 0 {
		return nil, fmt.Errorf("User not found")
	}

	var user models.User
	//,_
	user.ID = val["id"]
	user.Username = val["username"]
	user.Name = val["name"]
	user.Surname = val["surname"]

	return &user, nil
}

func (h *Handler) getUserInfo(key, field string) (string, error) {
	// key := fmt.Sprintf("user%d", id)
	val, err := h.client.HGet(key, field).Result()
	if err != nil {
		log.Fatal(err)
	}
	if len(val) == 0 {
		return "", fmt.Errorf("Value not found for field: %s", field)
	}
	return val, nil
}

// ! LOGIN USER HANDLER FUNCTIONS

func (h *Handler) login(username, password string) (*models.User, error) {
	userID, err := h.getIDByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving user ID")
	}

	key := fmt.Sprintf("user%s", userID)
	hashedPass, err := h.client.HGet(key, "password").Result()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password)); err != nil {
		return nil, fmt.Errorf("Incorrect password")
	}

	_, err = h.client.Set("currentUserId", userID, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("Error setting current user ID: %v", err)
	}

	user, err := h.getUserByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ! UPDATE USER HANDLER FUNCTIONS
func (h *Handler) update(userInfo *models.User) (*models.User, error) {

	Id, err := h.client.Get("currentUserId").Result()
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("user%s", Id)
	oldUsername, err := h.client.HGet(key, "username").Result()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving old username: %w", err)
	}

	if userInfo.Username != "" {
		usernameKey := fmt.Sprintf("username:%s", userInfo.Username)
		_, err := h.client.Get(usernameKey).Result()
		if err == redis.Nil {
			h.client.Set(usernameKey, Id, 0)
			h.client.HSet(key, "username", userInfo.Username)
			oldKey := fmt.Sprintf("username:%s", oldUsername)
			h.client.Del(oldKey)
		} else if err != nil {
			return nil, fmt.Errorf("Unknown error: %w", err)
		} else {
			return nil, fmt.Errorf("Username already exists")
		}
	}

	if userInfo.Password != "" {
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("Error hashing password: %w", err)
		}
		h.client.HSet(key, "password", string(hashedPass))
	}

	if userInfo.Name != "" {
		h.client.HSet(key, "name", userInfo.Name)
	}

	if userInfo.Surname != "" {
		h.client.HSet(key, "surname", userInfo.Surname)
	}

	updatedUser, err := h.getUserByID(Id)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving updated user: %w", err)
	}

	return updatedUser, nil
}

//! SECTION: SIMULATION HANDLER FUNCTIONS

func (h *Handler) matchSimulation() error {
	lastUserID, err := h.client.Get("user_id").Result()
	if err != nil {
		return fmt.Errorf("User count error: %v", err)
	}

	lastID, err := strconv.Atoi(lastUserID)
	if err != nil {
		return fmt.Errorf("Conversion error: %v", err)
	}
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
	for i := 1; i < lastID; i++ {
		for j := i + 1; j <= lastID; j++ {
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

	log.Printf("All players got matched.")
	return nil
}

func (h *Handler) createSimUser() error {
	const defaultPassword = "123456"

	newUserID := h.nextID()
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
	if err := h.storeUser(&simUser); err != nil {
		return err
	}
	return nil
}

// !SECTION: MATCH USERS HANDLER.
func (h *Handler) updateScore(match models.MatchInfo) error {
	firstUserKey := fmt.Sprintf("user%d", match.FirstUserId)
	secondUserKey := fmt.Sprintf("user%d", match.SecondUserId)

	if exists, err := h.client.Exists(firstUserKey).Result(); err != nil || exists == 0 {
		return errors.New("First user does not exist")
	}

	if exists, err := h.client.Exists(secondUserKey).Result(); err != nil || exists == 0 {
		return errors.New("Second user does not exist")
	}

	if match.FirstUserScore > match.SecondUserScore {
		if err := h.incScore(firstUserKey, 3); err != nil {
			return err
		}
	} else if match.FirstUserScore < match.SecondUserScore {
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
	for _, user := range leaderboard {
		rank, err := h.client.ZRevRank("leaderboard", user).Result()
		id, err := h.getUserInfo(user, "id")
		if err != nil {
			return nil, err
		}
		username, err := h.getUserInfo(user, "username")
		if err != nil {
			return nil, err
		}

		userInfo := map[string]interface{}{
			"rank":     rank + 1,
			"id":       id,
			"username": username,
		}
		list = append(list, userInfo)
	}
	return list, nil
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
	errorResponse := models.ErrorRespons{Status: false, Message: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}
func successResponse(w http.ResponseWriter, result models.SuccessRespons) {
	result.Status = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)
}

//!
/*
func (h *Handler) getUserScore(id int) (int, error) {
	key := fmt.Sprintf("user%d", id)
	val, err := h.client.HGet(key, "score").Result()
	if err != nil {
		return 0, fmt.Errorf("Could not get score: %v", err)
	}
	if val == "" {
		return 0, fmt.Errorf("Score not found")
	}

	score, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("Could not convert score to integer: %v", err)
	}

		return score, nil
	}
}
*/
