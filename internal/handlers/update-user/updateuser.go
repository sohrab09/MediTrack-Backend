package updateuser

import (
	"database/sql"
	"encoding/json"
	"log"
	"meditrack-backend/internal/models"
	"net/http"
	"strconv"
)

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()

		id := r.PathValue("id")

		userID, err := strconv.Atoi(id)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid user ID",
			})
			return
		}

		var user models.User

		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid JSON format",
			})
			return
		}

		result, err := db.Exec(
			`UPDATE users 
			 SET firstName=$1,lastName=$2,phone=$3,email=$4,status=$5,role=$6 
			 WHERE id=$7`,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.Email,
			user.Status,
			user.Role,
			userID,
		)

		if err != nil {
			log.Println("Update user error:", err)

			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Database error",
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Database error",
			})
			return
		}

		if rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "User not found",
			})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "User updated successfully",
		})
	}
}
