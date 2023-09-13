package main

import (
	"log"
	"net/http"

	api "github.com/Dzdrgl/redis-Api/api"
	"github.com/go-redis/redis"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	apiHandler := api.NewHandler(client)

	//? USER TRANSACTIONS
	http.HandleFunc("/v2/users/leaderboard", apiHandler.FetchLeaderboardPage)
	http.HandleFunc("/v2/users/{id:[0-9]+}", apiHandler.RetrieveUserByID)
	http.HandleFunc("/v2/users/new", apiHandler.CreateUser)
	http.HandleFunc("/v2/users/update", apiHandler.UpdateUser)
	http.HandleFunc("/v2/users/login", apiHandler.UserLogin)

	//? MATCH INFO
	http.HandleFunc("/v2/match", apiHandler.GetMatchInfo)

	//? SIMULATOR
	http.HandleFunc("/v2/simulator", apiHandler.Simulation)

	log.Println("Server is running on port 9090")
	http.ListenAndServe(":9090", nil)
}
