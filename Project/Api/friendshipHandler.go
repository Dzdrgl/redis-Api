package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Dzdrgl/redis-Api/models"
	"github.com/go-redis/redis"
)

type FriendRequestStatusType int

const (
	Accept FriendRequestStatusType = iota
	Reject
)

var friendRequestStatusTypeStrings = []string{"Accept", "Reject"}

func (s FriendRequestStatusType) String() string {
	return friendRequestStatusTypeStrings[s]
}

func (s *FriendRequestStatusType) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	for i, v := range friendRequestStatusTypeStrings {
		if v == str {
			*s = FriendRequestStatusType(i)
			return nil
		}
	}
	return errors.New("invalid status value")
}

func (h *Handler) HandleSearchUser(w http.ResponseWriter, r *http.Request) {
	log.Println("SearchUser - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	var requestUser models.User
	if err := json.NewDecoder(r.Body).Decode(&requestUser); err != nil {
		log.Printf("HandleSearchUser - Invalid JSON input: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}

	user, ok := r.Context().Value("userInfo").(models.User)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User info not found in context")
		return
	}

	if user.Username == requestUser.Username {
		errorResponse(w, http.StatusBadRequest, "Can't search your own username")
		return
	}

	searchedID := h.GetUserIDWithUsername(requestUser.Username)
	if searchedID == "" {
		log.Printf("HandleSearchUser - Username does not exist: %v", requestUser.Username)
		errorResponse(w, http.StatusBadRequest, "Username does not exist")
		return
	}

	successResponse(w, models.SuccessResponse{Status: true, Result: map[string]interface{}{"id": searchedID}})
}

func (h *Handler) HandleSendFriendRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleSendFriendRequest - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	var requestUser models.User
	if err := json.NewDecoder(r.Body).Decode(&requestUser); err != nil {
		log.Printf("HandleSendFriendRequest - Invalid JSON input: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}

	currentUser, ok := r.Context().Value("userInfo").(models.User)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User info not found in context")
		return
	}

	if currentUser.ID == requestUser.ID {
		errorResponse(w, http.StatusBadRequest, "Can't send friend request to yourself.")
		return
	}

	targetUsername := h.FetchUserFieldWithID(requestUser.ID, "username")
	if targetUsername == "" {
		errorResponse(w, http.StatusBadRequest, "User ID does not exist")
		return
	}
	if err := h.SentRequest(currentUser, requestUser.ID); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	responseMessage := "Friend request sent successfully to " + targetUsername
	log.Printf("HandleSendFriendRequest - %s", responseMessage)
	successResponse(w, models.SuccessResponse{Status: true, Result: map[string]interface{}{"message": responseMessage}})
}

func (h *Handler) HandleRequestList(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleRequestList - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	var listInfo models.ListInfo
	if err := json.NewDecoder(r.Body).Decode(&listInfo); err != nil {
		log.Printf("HandleRequestList - Invalid JSON input: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}

	currentUser, ok := r.Context().Value("userInfo").(models.User)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User info not found in context")
		return
	}

	if listInfo.Count <= 0 || listInfo.Page <= 0 {
		errorResponse(w, http.StatusBadRequest, "Invalid page or count value")
		return
	}

	friendRequests, err := h.FetchFriendRequests(listInfo, currentUser.ID)
	if err != nil {
		log.Printf("HandleRequestList - Error fetching friend requests: %v", err)
		errorResponse(w, http.StatusInternalServerError, "Could not retrieve friend requests")
		return
	}

	successResponse(w, models.SuccessResponse{Status: true, Result: friendRequests})
}

func (h *Handler) FetchFriendRequests(listInfo models.ListInfo, userID string) ([]models.FriendRequest, error) {
	var friendRequests []models.FriendRequest

	startIndex := listInfo.Count * (listInfo.Page - 1)
	endIndex := startIndex + listInfo.Count - 1

	results, err := h.client.ZRangeWithScores("requests:"+userID, startIndex, endIndex).Result()
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		var request models.FriendRequest
		request.Id = result.Member.(string)
		request.Username = h.FetchUserFieldWithID(result.Member.(string), "username")

		request.Date = time.Unix(int64(result.Score), 0).Format("2006-01-02T15:04:05")
		friendRequests = append(friendRequests, request)
	}

	return friendRequests, nil
}

type FriendRequestStatus struct {
	ID     string                  `json:"id"`
	Status FriendRequestStatusType `json:"status"`
}

func (h *Handler) HandleFriendRequestResponse(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleFriendRequestResponse - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	var status FriendRequestStatus
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		log.Printf("HandleFriendRequestResponse - Invalid JSON input: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}

	currentUser, ok := r.Context().Value("userInfo").(models.User)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User info not found in context")
		return
	}

	if err := h.processFriendRequestResponse(status, currentUser); err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(w, models.SuccessResponse{Status: true, Result: "Request processed successfully"})
}
func (h *Handler) processFriendRequestResponse(status FriendRequestStatus, currentUser models.User) error {
	if err := h.removeFriendRequest(currentUser.ID, status.ID); err != nil {
		return err
	}

	switch status.Status {
	case Accept:
		return h.addFriends(currentUser.ID, status.ID)
	case Reject:
		return nil
	default:
		return errors.New("Invalid status value")
	}
}

func (h *Handler) removeFriendRequest(userID, requestId string) error {
	val, err := h.client.ZRem("requests:"+userID, requestId).Result()
	if err != nil {
		return err
	} else if val == 0 {
		return errors.New("No such request ID found")
	}
	return nil
}
func (h *Handler) addFriends(senderID, reciperID string) error {

	if err := h.addFriend(senderID, reciperID); err != nil {
		return err
	}
	if err := h.addFriend(reciperID, senderID); err != nil {
		return err
	}
	return nil
}
func (h *Handler) addFriend(senderID, reciperID string) error {

	val, err := h.client.ZAdd("friends:"+senderID, redis.Z{Member: reciperID}).Result()
	if err != nil {
		return err
	} else if val == 0 {
		return errors.New("Already friends")
	}
	return nil
}
func (h *Handler) HandleListFriends(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleListFriends - Called")
	w.Header().Set(ContentType, ApplicationJSON)

	var listInfo models.ListInfo
	if err := json.NewDecoder(r.Body).Decode(&listInfo); err != nil {
		log.Printf("HandleListFriends - Invalid JSON input: %v", err)
		errorResponse(w, http.StatusBadRequest, InvalidJSONInputMsg)
		return
	}

	currentUser, ok := r.Context().Value("userInfo").(models.User)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User info not found in context")
		return
	}
	if listInfo.Count == 0 || listInfo.Page == 0 {
		log.Printf("HandleListFriends - The number of pages and users should not be zero..")
		errorResponse(w, http.StatusNotFound, "The number of pages and users should not be zero.")
		return
	}
	list, err := h.BuildFriendList(listInfo, currentUser.ID)
	if err != nil {
		log.Println("Error building list:", err)
		errorResponse(w, http.StatusNotFound, "Could not build list")
		return
	}
	result := models.SuccessResponse{
		Status: true,
		Result: list,
	}
	successResponse(w, result)
}

func (h *Handler) BuildFriendList(listInfo models.ListInfo, id string) ([]models.User, error) {
	var friendList []models.User
	startIndex := listInfo.Count * (listInfo.Page - 1)
	endIndex := startIndex + listInfo.Count - 1
	results, err := h.client.ZRevRangeWithScores("friends:"+id, startIndex, endIndex).Result()
	if err != nil {
		return nil, err
	}
	for _, user := range results {
		var userInfo models.User
		userInfo.ID = user.Member.(string)
		userInfo.Username = h.FetchUserFieldWithID(user.Member.(string), "username")
		friendList = append(friendList, userInfo)
	}

	return friendList, nil
}
