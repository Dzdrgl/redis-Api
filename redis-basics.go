package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var client *redis.Client

// User struct is used for JSON serialization.
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

// UserInfo holds public-facing user details.
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}
type Result struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// SuccessRespons holds the structure for success messages.
type SuccessRespons struct {
	Status bool   `json:"status"`
	Result Result `json:"result"`
}

// ErrorRespons holds the structure for error messages.
type ErrorRespons struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// Main func
func main() {
	// Initialize Redis client
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})

	fmt.Println("Server is running on port 8080")
	http.HandleFunc("/api/login", getUserByID)
	http.HandleFunc("/users/", getUserByID)
	http.HandleFunc("/users", getUsers)
	http.HandleFunc("/users/new", createUser)
	http.ListenAndServe(":8080", nil)
}

// WriteJSONResponse writes a JSON response to the HTTP writer.

func WriteSuccessResponse(w http.ResponseWriter, result Result) {
	successRespons := SuccessRespons{Status: true, Result: result}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(successRespons)
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := ErrorRespons{Status: false, Message: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// getUserByID handles requests to get a user by their ID.
func getUserByID(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	response.Header().Set("Content-Type", "application/json")

	if request.Method != "GET" {
		WriteErrorResponse(response, 405, "Method not allowed")
		return
	}

	// Extract user ID from the URL path.
	idStr := request.URL.Path[len("/users/"):]
	// Convert the ID to integer.
	id, err := strconv.Atoi(idStr)

	if err != nil {
		WriteErrorResponse(response, 400, "Invalid ID")
		return
	}

	// Get user details from Redis.
	key := fmt.Sprintf("user:%d", id)
	val, err := client.Get(ctx, key).Result()

	if err != nil {
		WriteErrorResponse(response, 404, "User not found")
		return
	}

	var user User
	var userInfo UserInfo

	// Deserialize the user data.
	err = json.Unmarshal([]byte(val), &user)
	if err != nil {
		WriteErrorResponse(response, 500, "Server error")
		return
	}

	// Prepare public user info.
	userInfo.ID = user.ID
	userInfo.Username = user.Username
	userInfo.Name = user.Name
	userInfo.Surname = user.Surname

	// Serialize and send the public user info.
	json.NewEncoder(response).Encode(userInfo)
}

func getUsers(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	response.Header().Set("Content-Type", "application/json")

	if request.Method != "GET" {
		WriteErrorResponse(response, 405, "Method not allowed")
		return
	}

	userIDs, err := client.SMembers(ctx, "users").Result()
	if err != nil {
		WriteErrorResponse(response, 500, "Server error")
		return
	}
	if len(userIDs) == 0 {
		WriteErrorResponse(response, 404, "No users found")
		return
	}

	var users []UserInfo

	for _, idStr := range userIDs {
		key := fmt.Sprintf("user:%s", idStr)
		val, err := client.Get(ctx, key).Result()
		if err != nil {
			message := fmt.Sprintf("Error getting user with ID %s: %v", idStr, err)
			WriteErrorResponse(response, 404, message)
			continue
		}

		var user User
		var userInfo UserInfo

		err = json.Unmarshal([]byte(val), &user)
		if err != nil {
			message := fmt.Sprintf("Error unmarshalling user: %v", err)
			WriteErrorResponse(response, 404, message)
			continue
		}
		userInfo.ID = user.ID
		userInfo.Username = user.Username
		userInfo.Name = user.Name
		userInfo.Surname = user.Surname

		users = append(users, userInfo)
	}
	// Serialize and send the public user info.
	json.NewEncoder(response).Encode(users)

}

func createUser(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	response.Header().Set("Content-Type", "application/json")

	// POST request kontrolü
	if request.Method != "POST" {
		WriteErrorResponse(response, 405, "Method not allowed")
		return
	}

	// Request body'yi decode ediyoruz
	var newUser User
	err := json.NewDecoder(request.Body).Decode(&newUser)
	if err != nil {
		WriteErrorResponse(response, 400, "Bad Request")
		return
	}

	// Username ve Password boş olmamalı
	if newUser.Username == "" || newUser.Password == "" {
		WriteErrorResponse(response, 400, "Username / Password must not be empty")
		return
	}

	// Username'in benzersiz olup olmadığını kontrol et
	added, err := client.SAdd(ctx, "usernames", newUser.Username).Result()
	if err != nil || added == 0 {
		WriteErrorResponse(response, 400, "Username already exists")
		return
	}

	// Şifreyi hashliyoruz
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteErrorResponse(response, 400, "Password hashing failed")
		return
	}
	newUser.Password = string(hashedPassword)

	// Kullanıcı ID'si oluşturuyoruz
	id, err := client.Incr(ctx, "user_id").Result()
	if err != nil {
		WriteErrorResponse(response, 400, "Failed to generate user ID")
		return
	}
	newUser.ID = int(id)

	// Kullanıcıyı Redis'e kaydediyoruz
	jsonData, err := json.Marshal(newUser)
	if err != nil {
		WriteErrorResponse(response, 500, "Server error")
		return
	}
	key := fmt.Sprintf("user:%d", id)
	client.Set(ctx, key, jsonData, 0)
	client.SAdd(ctx, "users", fmt.Sprintf("%d", id))

	// Başarılı yanıtı gönderiyoruz
	WriteSuccessResponse(response, Result{ID: newUser.ID, Username: newUser.Username})
}
