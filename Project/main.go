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
	http.HandleFunc("/v2/users/leaderboard", apiHandler.HandleLeaderboard)
	http.HandleFunc("/v2/users/", apiHandler.HandleRetrieveUser)
	http.HandleFunc("/v2/users/new", apiHandler.HandleCreateUser)
	http.HandleFunc("/v2/users/update", apiHandler.HandleUpdateUser)
	http.HandleFunc("/v2/users/login", apiHandler.HandleUserLogin)

	//? MATCH INFO
	http.HandleFunc("/v2/match", apiHandler.HandleMatch)

	//? SIMULATOR
	http.HandleFunc("/v2/simulator", apiHandler.HandleSimulation)
	//?Friendship
	http.HandleFunc("/v2/users/search", apiHandler.HandlerSearchUser)

	log.Println("Server is running on port 9090")
	http.ListenAndServe(":9090", nil)
}
