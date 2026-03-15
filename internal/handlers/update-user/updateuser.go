// Create a update user handler
package updateuser

import (
	"database/sql"
	"encoding/json"
	"meditrack-backend/internal/models"
	"net/http"
)

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Method check
		if r.Method != http.MethodPut {
			respondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed",
			})
			return
		}

		// Extract `{id}` from URL path
		id := r.PathValue("id")
		if id == "" {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "User ID is required",
			})
			return
		}

		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid JSON format",
			})
			return
		}

		// Update user in DB
		_, err = db.Exec(
			"UPDATE users SET firstName=$1, lastName=$2, phone=$3, email=$4, status=$5, role=$6 WHERE id=$7",
			user.FirstName, user.LastName, user.Phone, user.Email, user.Status, user.Role, id,
		)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Database error",
			})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "User updated successfully",
		})
	}
}
