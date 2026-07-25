package models

import "time"

type MedicineLeaf struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	QtyPerLeaf  int       `json:"qty_per_leaf"`
	Description string    `json:"description,omitempty"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
