package models

import "time"

type RegisterRequest struct {
	Data struct {
		FirstName string    `json:"firstName"`
		LastName  string    `json:"lastName"`
		Email     string    `json:"email"`
		Phone     string    `json:"phone"`
		Password  string    `json:"password"`
		CreatedAt time.Time `json:"created_at"`
		Role      int       `json:"role"`
	} `json:"data"`
}
