package register

import (
	"database/sql"
	"encoding/json"
	"log"
	"meditrack-backend/internal/models"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()

		var req models.RegisterRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid JSON format",
			})
			return
		}

		data := req.Data

		if data.FirstName == "" || data.LastName == "" || data.Email == "" ||
			data.Phone == "" || data.Password == "" || data.Role == 0 {

			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "All fields are required",
			})
			return
		}

		if len(data.Password) < 6 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Password must be at least 6 characters",
			})
			return
		}

		status := 1
		now := time.Now()

		hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Hash error:", err)
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Server error",
			})
			return
		}

		query := `
		INSERT INTO users (firstName, lastName, phone, email, password, role, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
		`

		var id int

		err = db.QueryRow(
			query,
			data.FirstName,
			data.LastName,
			data.Phone,
			data.Email,
			string(hashed),
			data.Role,
			status,
			now,
		).Scan(&id)

		if err != nil {
			log.Println("Insert error:", err)

			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Email or phone already exists",
			})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "User registered successfully",
			"data": map[string]interface{}{
				"id":         id,
				"firstName":  data.FirstName,
				"lastName":   data.LastName,
				"phone":      data.Phone,
				"email":      data.Email,
				"role":       data.Role,
				"status":     status,
				"created_at": now,
			},
		})
	}
}
