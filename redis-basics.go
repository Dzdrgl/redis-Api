package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var client *redis.Client

// Json dönüşümü içi User objesine json etiketleri veriyoruz.
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

// Main func
func main() {
	//Redis erişmi için client adında bir referans oluşturuyoruz.
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis adresi
	})
	fmt.Println("Server is running on port 8080")

	http.HandleFunc("/users/", getUserByID)
	http.HandleFunc("/users", getUsers)
	http.HandleFunc("/users/new", createUser)
	http.ListenAndServe(":8080", nil)

}
func getUserByID(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	response.Header().Set("Content-Type", "application/json")

	if request.Method != "GET" {
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Pathdeki "/users/" kısmını çıkar geri kalan string türündeki veriyi al.
	idStr := request.URL.Path[len("/users/"):]
	//Burada  idStr ye atanan string ifadesinin tip dönüşümü yapılır. Değer int'e çevrilir.
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(response, "Invalid ID", http.StatusBadRequest)
		return
	}

	//Sprintf değişkenleri alır ve bunları birleştirerek bir string döndürür.
	key := fmt.Sprintf("user:%d", id)
	// Burada ctx iptal, zaman aşımı veya deadline bilgilerini taşır.
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		http.Error(response, "User not found", http.StatusNotFound)
		return
	}

	var user User
	var userInfo UserInfo
	//Redis'den alinan deger val a atianir ve burada olusturlan user degiseknin adresine atar

	err = json.Unmarshal([]byte(val), &user)
	if err != nil {
		http.Error(response, "Server error", http.StatusInternalServerError)
		return
	}
	if user.ID == 1 {
		json.NewEncoder(response).Encode(user)
	} else {

		userInfo.ID = user.ID
		userInfo.Username = user.Username
		userInfo.Name = user.Name
		userInfo.Surname = user.Surname

		json.NewEncoder(response).Encode(userInfo)
	}

}

func getUsers(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	response.Header().Set("Content-Type", "application/json")

	if request.Method != "GET" {
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDs, err := client.SMembers(ctx, "users").Result()
	if err != nil {
		http.Error(response, "Server error", http.StatusInternalServerError)
		return
	}

	var users []User

	for _, idStr := range userIDs {
		key := fmt.Sprintf("user:%s", idStr)
		val, err := client.Get(ctx, key).Result()
		if err != nil {
			log.Printf("Error getting user with ID %s: %v", idStr, err)
			continue
		}

		var user User
		err = json.Unmarshal([]byte(val), &user)
		if err != nil {
			log.Printf("Error unmarshalling user: %v", err)
			continue
		}
		users = append(users, user)
	}

	json.NewEncoder(response).Encode(users)
}

func createUser(response http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	response.Header().Set("Content-Type", "application/json")

	if request.Method != "POST" {
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newUser User
	err := json.NewDecoder(request.Body).Decode(&newUser)
	if err != nil {
		http.Error(response, "Bad Request", http.StatusBadRequest)
		return
	}
	//Kullanici adi ve sifre alanlarinin bos olup olmadiginin kontrolu yapiliyor.
	if newUser.Username == "" || newUser.Password == "" {
		http.Error(response, "Failed", http.StatusBadRequest)
		json.NewEncoder(response).Encode(map[string]interface{}{
			"status":  false,
			"message": "Username and Password cannot be empty",
		})
		return
	}
	//Kullanici adi usernames keyi ile listeye ekleniyor. Eger result olarak 0 dönerse bu zaten listede bu değerde bir elemanın olduğu anlamına geliyor. Bu sayede username uniq bir yapıda oluyor.
	added, err := client.SAdd(ctx, "usernames", newUser.Username).Result()
	if err != nil {
		http.Error(response, "Failed", http.StatusInternalServerError)
		json.NewEncoder(response).Encode(map[string]interface{}{
			"status":  false,
			"message": "Server error",
		})
		return
	}
	if added == 0 {
		http.Error(response, "Failed", http.StatusInternalServerError)
		json.NewEncoder(response).Encode(map[string]interface{}{
			"status":  false,
			"message": "Username already exists",
		})
		return

	} else {
		//Diğer değer atama işlemlerini else bloğunun altında yapmamın sebebi işlemler başarızı olsa bile user_id değeri artıyor ve gereksizdeğer atamaları yapılıyor bu şekilde bnların önüne geçtim.

		//Kullanıcı şifrelerini  redise kaydedilmeden önce hashleniyor burada hashlemeyi bcrypt kütüphanesi ile yaptım.
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(response, "Failed", http.StatusInternalServerError)
			json.NewEncoder(response).Encode(map[string]interface{}{
				"status":  false,
				"message": "Error hashing the password",
			})
			return
		}
		//Kulanıcının şifresi hashlenmiş olarak kullanıcının şifrsi ile değiştriliyor
		newUser.Password = string(hashedPassword)

		//Kullanıcıya id atamasının bir arttırılıp yapılması.
		newID, err := client.Incr(ctx, "user_id").Result()
		if err != nil {
			http.Error(response, "Failed", http.StatusInternalServerError)
			json.NewEncoder(response).Encode(map[string]interface{}{
				"status":  false,
				"message": "Server error",
			})
			return
		}
		//Dönen iddeğerinin kullanıcıya atanması.
		newUser.ID = int(newID)

		//Oluşan User objesinin json formatına uygun pars edilmesi işlemi.
		jsonData, err := json.Marshal(newUser)
		if err != nil {
			http.Error(response, "Server error", http.StatusInternalServerError)
			return
		}

		//json datayı sitring olarak SET ile user:id keyi ile redise kayıt etme işlemi.
		key := fmt.Sprintf("user:%d", newID)
		client.Set(ctx, key, jsonData, 0)
		client.SAdd(ctx, "users", fmt.Sprintf("%d", newID))

		// Oluşan kullancının kayıt işlemi başarılı olduğunda ekrana kayıt olan kulanıcın bilgilierinin json formatında gösterilmesi.
		response.WriteHeader(http.StatusCreated)
		json.NewEncoder(response).Encode(map[string]interface{}{
			"status": true,
			"result": map[string]interface{}{
				"id":       newUser.ID,
				"username": newUser.Username,
			},
		})
	}
}
