package models

import "time"

type MedicineBoxSize struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TotalPcs  int       `json:"total_pcs"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
