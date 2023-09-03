package main

import (
	"fmt"
	"net/http"

	api "github.com/Dzdrgl/redis-Api/Api"
	"github.com/go-redis/redis"
)

func main() {
	api.Client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	http.HandleFunc("/v2/users/", api.GetUserByID)
	http.HandleFunc("/v2/users/update", api.UpdateUser)
	http.HandleFunc("/v2/users/new", api.CreateUser)
	http.HandleFunc("/v2/users/login", api.UserLogin)
	fmt.Println("Server is running on port 9090")
	http.ListenAndServe(":9090", nil)
}
