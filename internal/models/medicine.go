package models

import "time"

type Medicine struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Generic      *string   `json:"generic,omitempty"`
	Strength     *string   `json:"strength,omitempty"`
	UnitID       int       `json:"unit_id"`
	CategoryID   int       `json:"category_id"`
	LeafID       *int      `json:"leaf_id,omitempty"`
	TypeID       *int      `json:"type_id,omitempty"`
	BoxSizeID    int       `json:"box_size_id"`
	MinimumStock int       `json:"minimum_stock"`
	CurrentStock int       `json:"-"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
