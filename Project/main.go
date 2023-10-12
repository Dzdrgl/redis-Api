package main

import (
	"log"
	"net/http"

	"github.com/Dzdrgl/redis-Api/api"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	handler := api.NewHandler(rdb)

	//? User routes
	router.HandleFunc("/api/v2/users/{id:[0-9]+}", handler.AuthMiddleware(handler.HandleRetrieveUser)).Methods("GET")
	router.HandleFunc("/api/v2/users/new", handler.HandleCreateUser).Methods("POST")
	router.HandleFunc("/api/v2/users/update", handler.AuthMiddleware(handler.HandleUpdateUser)).Methods("PUT")
	router.HandleFunc("/api/v2/users/login", handler.HandleUserLogin).Methods("POST")

	router.HandleFunc("/api/v2/users/leaderboard", handler.AuthMiddleware(handler.HandleLeaderboard)).Methods("POST")
	//? MATCH INFO
	router.HandleFunc("/api/v2/match", handler.HandleMatch).Methods("POST")

	//? SIMULATOR
	router.HandleFunc("/api/v2/simulator", handler.HandleSimulation)
	//?Friendship
	router.HandleFunc("/api/v2/users/search", handler.AuthMiddleware(handler.HandleSearchUser)).
		Methods("POST")
	router.HandleFunc("/api/v2/users/sent", handler.AuthMiddleware(handler.HandleSendFriendRequest)).Methods("POST")
	router.HandleFunc("/api/v2/users/requests", handler.AuthMiddleware(handler.HandleRequestList)).Methods("POST")
	router.HandleFunc("/api/v2/users/requests/status", handler.AuthMiddleware(handler.HandleFriendRequestResponse)).Methods("POST")
	router.HandleFunc("/api/v2/users/friends", handler.AuthMiddleware(handler.HandleListFriends)).Methods("POST")
	// Start the server
	http.Handle("/", router)
	log.Println("Server is running on port 9090")
	http.ListenAndServe(":9090", nil)
}

// htmlContent, err := ioutil.ReadFile(htmlFileName)
// 	if err != nil {
// 		log.Fatalf("HTML dosyasını okurken hata oluştu: %v", err)
// 	}
// 	router.HandleFunc("/api/v2/info", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, string(htmlContent))
// 	}).Methods("GET")
