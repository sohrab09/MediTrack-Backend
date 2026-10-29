package models

type LoginRequest struct {
	Data LoginCredentials `json:"data"`
}

type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
