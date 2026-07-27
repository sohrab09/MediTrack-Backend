package models

import "time"

type Medicine struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Strength   string    `json:"strength,omitempty"`
	Generic    string    `json:"generic,omitempty"`
	CategoryID int       `json:"category_id"`
	TypeID     *int      `json:"type_id,omitempty"`
	BoxSizeID  int       `json:"box_size_id"`
	UnitID     int       `json:"unit_id"`
	LeafID     *int      `json:"leaf_id,omitempty"`
	Price      float64   `json:"price"`
	Discount   float64   `json:"discount"`
	Tax        float64   `json:"tax"`
	Vat        float64   `json:"vat"`
	Status     int       `json:"status"`
	ImageURL   *string   `json:"image_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
