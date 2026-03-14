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

// Helper to send JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "JSON encode error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Write(response)
}

// Register Handler
func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Method check
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed",
			})
			return
		}

		// Decode request
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid JSON format",
			})
			return
		}

		// Taking input
		data := req.Data

		if data.FirstName == "" || data.LastName == "" || data.Email == "" || data.Phone == "" || data.Password == "" || data.Role == 0 {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "All fields are required",
			})
			return
		}

		// Default status
		status := 1 // active

		// Password hashing
		hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Password hashing error:", err)
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Server error",
			})
			return
		}

		// DB Query
		query := `INSERT INTO users (firstName, lastName, phone, email, password, role, status, created_at)
          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

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
			time.Now(),
		).Scan(&id)

		if err != nil {
			log.Println("❌ Insert error:", err)
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Email or phone already exists",
				"status":  http.StatusBadRequest,
			})
			return
		}

		// Send response
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "User registered successfully",
			"status":  http.StatusOK,
			"data": map[string]interface{}{
				"id":         id,
				"firstName":  data.FirstName,
				"lastName":   data.LastName,
				"phone":      data.Phone,
				"email":      data.Email,
				"role":       data.Role,
				"status":     status,
				"created_at": time.Now(),
			},
		})
	}
}
