package login

import (
	"database/sql"
	"encoding/json"
	"log"
	"meditrack-backend/internal/models"
	"meditrack-backend/internal/utils"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Helper to send JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// Login Handler
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Method check
		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		// Decode request
		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Invalid JSON format.",
			})
			return
		}

		// Taking input
		var user models.User
		var hashed string
		err := db.QueryRow(
			`SELECT id, firstName, lastName, phone, email, password, role, status, created_at 
			 FROM users 
			 WHERE email=$1`,
			req.Data.Email,
		).Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Phone,
			&user.Email,
			&hashed,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
		)

		// Error check
		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid Credentials.",
			})
			return
		}
		if err != nil {
			log.Println("DB query err", err)
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Server error.",
			})
			return
		}

		// Password check
		if bcrypt.CompareHashAndPassword([]byte(hashed), []byte(req.Data.Password)) !=
			nil {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid Credentials.",
			})
			return
		}
		// Generate JWT
		token, err := utils.GenerateJWT(user.Email)
		if err != nil {
			log.Println("JWT error:", err)
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Error generating token",
			})
			return
		}

		// Send response
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Login successful",
			"data": map[string]interface{}{
				"user":  user,
				"token": token,
			},
		})
	}
}
