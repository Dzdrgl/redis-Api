package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	models "github.com/Dzdrgl/redis-Api/Models"
	utilities "github.com/Dzdrgl/redis-Api/Utilities"
	"github.com/go-redis/redis"
	"golang.org/x/crypto/bcrypt"
)

var ErrorRespons func(w http.ResponseWriter, statusCode int, message string) = utilities.WriteErrorResponse

var SuccessRespons func(w http.ResponseWriter, result models.SuccessRespons) = utilities.WriteSuccessResponse

var Client *redis.Client

const (
	emptyErr         = "Username/Password is empty"
	existsErr        = "Username already exists"
	usernameNotFound = "Username doesn't exist"
	wrongPassword    = "Wrong Password"
)

func ValidateUser(newUser *models.User) error {
	if newUser.Username == "" || newUser.Password == "" {
		return fmt.Errorf(emptyErr)
	}
	if UsernameExists(newUser.Username) {
		return fmt.Errorf(existsErr)
	}
	return nil
}

func UsernameExists(username string) bool {
	exists, err := Client.SIsMember("usernames", username).Result()
	return err == nil && exists
}

func StoreUser(newUser *models.User) error {
	newUser.ID = NextID()
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUser.Password = string(hashedPass)
	jsonData, err := json.Marshal(newUser)
	if err != nil {
		return err
	}

	keyID := fmt.Sprintf("user:%d", newUser.ID)
	keyUsername := fmt.Sprintf("user:%s", newUser.Username)

	pipe := Client.Pipeline()
	pipe.Set(keyID, jsonData, 0)
	pipe.SAdd("usernames", newUser.Username)
	pipe.Set(keyUsername, newUser.ID, 0)

	_, err = pipe.Exec()
	return err
}
func SearchByUserID(id int) (models.User, error) {
	var user models.User
	key := fmt.Sprintf("user:%d", id)
	val, err := Client.Get(key).Result()
	if err != nil {
		return user, err
	}
	err = json.Unmarshal([]byte(val), &user)
	if err != nil {
		return user, err
	}
	return user, nil
}
func GetUserID(username string) (int, error) {
	key := fmt.Sprintf("user:%s", username)
	val, err := Client.Get(key).Result()
	id, err := strconv.Atoi(val)
	if err != nil {
		return id, err
	}
	return id, nil
}

func Login(username, password string) (models.User, error) {
	if !UsernameExists(username) {
		return models.User{}, fmt.Errorf("Username not found")
	}
	user_id, err := GetUserID(username)
	if err != nil {
		return models.User{}, fmt.Errorf("Error retrieving user ID")
	}
	user, err := SearchByUserID(user_id)
	if err != nil {
		return models.User{}, fmt.Errorf("Error retrieving user details")
	}
	if err != nil {
		return models.User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return models.User{}, fmt.Errorf("Incorrect password")
	}
	return user, nil
}

func NextID() int {
	userId, err := Client.Incr("user_id").Result()
	if err != nil {
		return 0
	}
	return int(userId)
}
