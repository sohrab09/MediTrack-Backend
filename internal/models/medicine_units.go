package models

import "time"

type MedicineUnits struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
