package models

type SuccessResponse struct {
	Status bool        `json:"status"`
	Result interface{} `json:"result"`
}
type ErrorResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type MatchInfo struct {
	FirstUserId     int `json:"firstuserid"`
	SecondUserId    int `json:"seconduserid"`
	FirstUserScore  int `json:"firstuserscore"`
	SecondUserScore int `json:"seconduserscore"`
}

type LeaderbordInfo struct {
	Count int64 `json:"count"`
	Page  int64 `json:"page"`
}

type SimulationInfo struct {
	Usercount int `json:"usercount"`
}
