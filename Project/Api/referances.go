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

func ValidateUser(newUser *models.User) error {
	if newUser.Username == "" || newUser.Password == "" {
		return fmt.Errorf("Username/Password is empty")
	}
	if UsernameExists(newUser.Username) {
		return fmt.Errorf("Username already exists")
	}
	return nil
}

func UsernameExists(username string) bool {
	exists, err := Client.SIsMember("usernames", username).Result()
	return err == nil && exists
}

func StoreUser(newUser *models.User) error {
	newUser.ID = NextID()
	if newUser.ID == 0 {
		return fmt.Errorf("UserId  == 0")
	}
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
	keyUsername := fmt.Sprintf("username:%s", newUser.Username)

	pipe := Client.Pipeline()
	pipe.Set(keyID, jsonData, 0)
	pipe.Set(keyUsername, newUser.ID, 0)
	pipe.SAdd("usernames", newUser.Username)

	cmders, err := pipe.Exec()
	if err != nil {
		return err
	}

	for _, cmder := range cmders {
		if cmder.Err() != nil {
			return fmt.Errorf("Pipeline command failed: %v", cmder.Err())
		}
	}
	return nil

}
func SearchByUserID(id int) (*models.User, error) {
	key := fmt.Sprintf("user:%d", id)
	return getUser(key)
}
func GetIDByUsername(username string) (int, error) {
	key := fmt.Sprintf("username:%s", username)
	val, err := Client.Get(key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func Login(username, password string) (*models.User, error) {
	if !UsernameExists(username) {
		return nil, fmt.Errorf("Username not found")
	}
	userID, err := GetIDByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving user ID")
	}
	user, err := SearchByUserID(userID)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("Incorrect password")
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

func getUser(key string) (*models.User, error) {
	var user models.User
	val, err := Client.Get(key).Result()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving user details: %v", err)
	}
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, fmt.Errorf("Error unmarshaling user data: %v", err)
	}
	return &user, nil
}

func Update(userInfo *models.User) (*models.User, error) {
	val, err := Client.Get("currentUser").Result()
	if err != nil {
		return nil, fmt.Errorf("No logged-in user found")
	}

	var currentUser models.User
	if err := json.Unmarshal([]byte(val), &currentUser); err != nil {
		return nil, fmt.Errorf("Error unmarshaling current user data: %v", err)
	}

	if userInfo.Username != "" && userInfo.Username != currentUser.Username {
		if UsernameExists(userInfo.Username) {
			return nil, fmt.Errorf("Username already exists")
		}
		currentUser.Username = userInfo.Username
	}

	if userInfo.Password != "" {
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("Error hashing password: %v", err)
		}
		currentUser.Password = string(hashedPass)
	}
	if userInfo.Name != "" {
		currentUser.Name = userInfo.Name
	}

	if userInfo.Surname != "" {
		currentUser.Surname = userInfo.Surname
	}

	jsonData, err := json.Marshal(currentUser)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling user to JSON")
	}

	key := fmt.Sprintf("user:%d", currentUser.ID)
	if err := Client.Set(key, jsonData, 0).Err(); err != nil {
		return nil, fmt.Errorf("Error updating user in Redis")
	}
	return &currentUser, nil
}
