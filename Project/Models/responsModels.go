package models

type SuccessResponse struct {
	Status bool        `json:"status"`
	Result interface{} `json:"result"`
}
type ErrorResponse struct {
	ErrorMessage string `json:"message"`
}

type MatchInfo struct {
	FirstUserId     int `json:"firstuserid"`
	SecondUserId    int `json:"seconduserid"`
	FirstUserScore  int `json:"firstuserscore"`
	SecondUserScore int `json:"seconduserscore"`
}

type ListInfo struct {
	Count int64 `json:"count"`
	Page  int64 `json:"page"`
}

type SimulationInfo struct {
	Usercount int `json:"usercount"`
}
type FriendRequest struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Date     string `json:"date"`
}
