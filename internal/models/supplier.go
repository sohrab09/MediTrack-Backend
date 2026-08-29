package models

import "time"

type Supplier struct {
	ID             int       `json:"id"`
	SupplierName   string    `json:"supplier_name"`
	Mobile         string    `json:"mobile"`
	Email          *string   `json:"email,omitempty"`
	ContactPerson  *string   `json:"contact_person,omitempty"`
	Address        *string   `json:"address,omitempty"`
	City           *string   `json:"city,omitempty"`
	State          *string   `json:"state,omitempty"`
	Zip            *string   `json:"zip,omitempty"`
	Country        string    `json:"country"`
	OpeningBalance float64   `json:"opening_balance"`
	CurrentBalance float64   `json:"current_balance"`
	Status         int       `json:"status"`
	CreatedBy      *int      `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
