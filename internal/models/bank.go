package models

import "time"

type Bank struct {
	ID                int       `json:"id"`
	BankName          string    `json:"bank_name"`
	AccountHolderName string    `json:"account_holder_name"`
	AccountNumber     string    `json:"account_number"`
	AccountType       string    `json:"account_type"`
	BranchName        string    `json:"branch_name"`
	RoutingNumber     string    `json:"routing_number"`
	OpeningBalance    float64   `json:"opening_balance"`
	CurrentBalance    float64   `json:"current_balance"`
	Status            int       `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
