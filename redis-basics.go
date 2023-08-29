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
func WriteJSONResponse(w http.ResponseWriter, r ErrorRespons) {
	jsonData, _ := json.Marshal(r)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// getUserByID handles requests to get a user by their ID.
func getUserByID(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	response.Header().Set("Content-Type", "application/json")

	var errorResponse ErrorRespons
	if request.Method != "GET" {
		errorResponse.Status = false
		errorResponse.Message = "Method not allowed"
		response.WriteHeader(405)
		WriteJSONResponse(response, errorResponse)
		return
	}

	// Extract user ID from the URL path.
	idStr := request.URL.Path[len("/users/"):]
	// Convert the ID to integer.
	id, err := strconv.Atoi(idStr)

	if err != nil {
		errorResponse.Status = false
		errorResponse.Message = "Invalid ID"
		response.WriteHeader(400)
		WriteJSONResponse(response, errorResponse)
		return
	}

	// Get user details from Redis.
	key := fmt.Sprintf("user:%d", id)
	val, err := client.Get(ctx, key).Result()

	if err != nil {
		errorResponse.Status = false
		errorResponse.Message = "User not found"
		response.WriteHeader(404)
		WriteJSONResponse(response, errorResponse)
		return
	}

	var user User
	var userInfo UserInfo

	// Deserialize the user data.
	err = json.Unmarshal([]byte(val), &user)
	if err != nil {
		errorResponse.Status = false
		errorResponse.Message = "Server error"
		response.WriteHeader(500)
		WriteJSONResponse(response, errorResponse)
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

	var errorResponse ErrorRespons

	var errCount int = 0

	if request.Method != "GET" {
		errCount++
		errorResponse.Status = false
		errorResponse.Message = "Method not allowed"
		response.WriteHeader(405)
		WriteJSONResponse(response, errorResponse)
		return
	}

	userIDs, err := client.SMembers(ctx, "users").Result()
	if err != nil {
		errCount++
		return
	}

	var users []UserInfo
	for _, idStr := range userIDs {
		key := fmt.Sprintf("user:%s", idStr)
		val, err := client.Get(ctx, key).Result()
		if err != nil {
			errCount++
			message := fmt.Sprintf("Error getting user with ID %s: %v", idStr, err)
			errorResponse.Status = false
			errorResponse.Message = message
			response.WriteHeader(404)
			WriteJSONResponse(response, errorResponse)
			continue
		}

		var user User
		var userInfo UserInfo

		err = json.Unmarshal([]byte(val), &user)
		if err != nil {
			errCount++
			message := fmt.Sprintf("Error unmarshalling user: %v", err)
			errorResponse.Status = false
			errorResponse.Message = message
			response.WriteHeader(404)
			WriteJSONResponse(response, errorResponse)
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

	var errorResponse ErrorRespons
	var successRespons SuccessRespons

	var errCount int = 0

	if request.Method != "POST" {
		errCount++
		errorResponse.Status = false
		errorResponse.Message = "Method not allowed"
		response.WriteHeader(405)
		WriteJSONResponse(response, errorResponse)
		return
	}

	var newUser User
	err := json.NewDecoder(request.Body).Decode(&newUser)
	if err != nil {
		errCount++
		errorResponse.Status = false
		errorResponse.Message = "Bad Request"
		response.WriteHeader(400)
		WriteJSONResponse(response, errorResponse)
		return
	}
	//Kullanici adi ve sifre alanlarinin bos olup olmadiginin kontrolu yapiliyor.
	if newUser.Username == "" || newUser.Password == "" {
		errCount++
		errorResponse.Status = false
		errorResponse.Message = "Username / Password is not empty "
		response.WriteHeader(400)
		WriteJSONResponse(response, errorResponse)
		return
	}
	//Kullanici adi usernames keyi ile listeye ekleniyor. Eger result olarak 0 dönerse bu zaten listede bu değerde bir elemanın olduğu anlamına geliyor. Bu sayede username uniq bir yapıda oluyor.
	added, err := client.SAdd(ctx, "usernames", newUser.Username).Result()
	if err != nil {
		errCount++
		errorResponse.Status = false
		errorResponse.Message = "Username isn't added"
		response.WriteHeader(400)
		WriteJSONResponse(response, errorResponse)
		return
	}
	if added == 0 {
		errCount++
		errorResponse.Status = false
		errorResponse.Message = "Username exist"
		response.WriteHeader(400)
		WriteJSONResponse(response, errorResponse)
		return

	} else {
		//Diğer değer atama işlemlerini else bloğunun altında yapmamın sebebi işlemler başarızı olsa bile user_id değeri artıyor ve gereksizdeğer atamaları yapılıyor bu şekilde bnların önüne geçtim.

		//Kullanıcı şifrelerini  redise kaydedilmeden önce hashleniyor burada hashlemeyi bcrypt kütüphanesi ile yaptım.
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			errCount++
			errorResponse.Status = false
			errorResponse.Message = "Password didn't hash"
			response.WriteHeader(400)
			WriteJSONResponse(response, errorResponse)
			return
		}

		//Kulanıcının şifresi hashlenmiş olarak kullanıcının şifrsi ile değiştriliyor
		newUser.Password = string(hashedPassword)

		//Kullanıcıya id atamasının bir arttırılıp yapılması.
		id, err := client.Incr(ctx, "user_id").Result()

		if err != nil {
			errCount++
			errorResponse.Status = false
			errorResponse.Message = ""
			response.WriteHeader(400)
			WriteJSONResponse(response, errorResponse)
			return
		} else {
			newUser.ID = int(id)
		}

		//Oluşan User objesinin json formatına uyguns hale getriilmes'.
		jsonData, err := json.Marshal(newUser)
		if err != nil {
			errCount++
			errorResponse.Status = false
			errorResponse.Message = "Server error"
			response.WriteHeader(500)
			WriteJSONResponse(response, errorResponse)
			return
		}

		//json datayı sitring olarak SET ile user:id keyi ile redise kayıt etme işlemi.
		key := fmt.Sprintf("user:%d", id)
		client.Set(ctx, key, jsonData, 0)
		client.SAdd(ctx, "users", fmt.Sprintf("%d", id))

		// Oluşan kullancının kayıt işlemi başarılı olduğunda ekrana kayıt olan kulanıcın bilgilierinin json formatında gösterilmesi.
		response.WriteHeader(http.StatusCreated)

		successRespons.Status = true
		successRespons.Result.ID = newUser.ID
		successRespons.Result.Username = newUser.Username
		json.NewEncoder(response).Encode(successRespons)

	}
}
