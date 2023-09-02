package models

type SuccessRespons struct {
	Status bool        `json:"status"`
	Result interface{} `json:"result"`
}

type ErrorRespons struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}
