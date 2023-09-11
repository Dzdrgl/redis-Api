package models

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
	Surname  string `json:"surname,omitempty"`
	Score    string `json:"score,omitempty"`
}

//omitempty
