package models

import "time"

type Medicine struct {
	ID           int       `json:"id"`
	Code         string    `json:"code"`
	Barcode      *string   `json:"barcode,omitempty"`
	Name         string    `json:"name"`
	Strength     *string   `json:"strength,omitempty"`
	Generic      *string   `json:"generic,omitempty"`
	CategoryID   int       `json:"category_id"`
	TypeID       *int      `json:"type_id,omitempty"`
	BoxSizeID    int       `json:"box_size_id"`
	UnitID       int       `json:"unit_id"`
	LeafID       *int      `json:"leaf_id,omitempty"`
	SellingPrice float64   `json:"selling_price"`
	MRP          float64   `json:"mrp"`
	CurrentStock int       `json:"current_stock"`
	MinimumStock int       `json:"minimum_stock"`
	Status       int       `json:"status"`
	CreatedBy    *int      `json:"created_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
