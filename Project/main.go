package main

import (
	"fmt"
	"net/http"

	api "github.com/Dzdrgl/redis-Api/api"
	"github.com/go-redis/redis"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	apiHandler := api.NewHandler(client)

	http.HandleFunc("/v2/users/leaderboard", apiHandler.Leaderboard)
	http.HandleFunc("/v2/match", apiHandler.GetMatchInfo)
	http.HandleFunc("/v2/users/", apiHandler.GetUserByID)
	http.HandleFunc("/v2/users/update", apiHandler.UpdateUser)
	http.HandleFunc("/v2/users/new", apiHandler.CreateUser)
	http.HandleFunc("/v2/users/login", apiHandler.UserLogin)
	http.HandleFunc("/v2/simulator", apiHandler.Simulator)
	fmt.Println("Server is running on port 9090")
	http.ListenAndServe(":9090", nil)
}
