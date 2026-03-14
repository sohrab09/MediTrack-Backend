package getusers

import (
	"database/sql"
	"encoding/json"
	"log"
	"meditrack-backend/internal/models"
	"net/http"
)

// respondJSON safely writes JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println("Failed to write JSON response:", err)
	}
}

func GetUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Method check
		if r.Method != http.MethodGet {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed",
			})
			return
		}

		// Query specific columns, excluding password
		rows, err := db.Query("SELECT id, firstName, lastName, phone, email, status, role, created_at FROM users")
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Database error",
			})
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var user models.User
			if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Phone, &user.Email, &user.Status, &user.Role, &user.CreatedAt); err != nil {
				respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Database error",
				})
				return
			}
			users = append(users, user)
		}

		// Check for rows iteration errors
		if err = rows.Err(); err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Error reading rows",
			})
			return
		}

		// Success response
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    users,
			"count":   len(users),
		})
	}
}
