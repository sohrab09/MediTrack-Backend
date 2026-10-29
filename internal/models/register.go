package models

type RegisterRequest struct {
	Data struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Password  string `json:"password"`
		Role      int    `json:"role"`
	} `json:"data"`
}
