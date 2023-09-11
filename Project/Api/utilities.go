package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis"

	models "github.com/Dzdrgl/redis-Api/models"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	client *redis.Client
}

func NewHandler(client *redis.Client) *Handler {
	return &Handler{client: client}
}

var errorResponse func(w http.ResponseWriter, statusCode int, message string) = writeErrorResponse
var successResponse func(w http.ResponseWriter, result models.SuccessRespons) = writeSuccessResponse

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

func (h *Handler) storeUser(newUser *models.User) error {
	newUser.ID = h.nextID()
	if newUser.ID == "" {
		return fmt.Errorf("UserId == 0")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUser.Password = string(hashedPass)

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
	_, err = pipe.Exec()
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) getIDByUsername(username string) (string, error) {
	key := fmt.Sprintf("username:%s", username)
	val, err := h.client.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (h *Handler) getUser(id string) (*models.User, error) {
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

	updatedUser, err := h.getUser(Id)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving updated user: %w", err)
	}

	return updatedUser, nil
}

//

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

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := models.ErrorRespons{Status: false, Message: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

func writeSuccessResponse(w http.ResponseWriter, result models.SuccessRespons) {
	result.Status = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)
}
