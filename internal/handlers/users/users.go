package users

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"meditrack-backend/internal/models"
)

// Standard API Response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Count   *int        `json:"count,omitempty"`
}

// respondJSON safely writes JSON response
func respondJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("Failed to write JSON response:", err)
	}
}

// Extract and validate ID from path parameter
func parseIDFromPath(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}

// GET /users - Fetch all users
func GetUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := `SELECT id, firstName, lastName, phone, email, status, role, created_at FROM users`
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("GetUsers query error: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Database error",
			})
			return
		}
		defer rows.Close()

		users := make([]models.User, 0)
		for rows.Next() {
			var user models.User
			if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Phone, &user.Email, &user.Status, &user.Role, &user.CreatedAt); err != nil {
				log.Printf("GetUsers scan error: %v", err)
				respondJSON(w, http.StatusInternalServerError, Response{
					Success: false,
					Message: "Database error",
				})
				return
			}
			users = append(users, user)
		}

		if err = rows.Err(); err != nil {
			log.Printf("GetUsers rows iteration error: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Error reading rows",
			})
			return
		}

		count := len(users)
		respondJSON(w, http.StatusOK, Response{
			Success: true,
			Data:    users,
			Count:   &count,
		})
	}
}

// GET /users/{id} - Fetch single user
func GetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := parseIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid or missing User ID",
			})
			return
		}

		query := `SELECT id, firstName, lastName, phone, email, status, role, created_at FROM users WHERE id=$1`

		var user models.User
		err = db.QueryRowContext(ctx, query, userID).
			Scan(&user.ID, &user.FirstName, &user.LastName, &user.Phone, &user.Email, &user.Status, &user.Role, &user.CreatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, Response{
				Success: false,
				Message: "User not found",
			})
			return
		}

		if err != nil {
			log.Printf("GetUser error: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Database error",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Success: true,
			Data:    user,
		})
	}
}

// PUT /users/{id} - Update user details
func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := parseIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid user ID",
			})
			return
		}

		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid JSON format",
			})
			return
		}

		query := `UPDATE users 
				  SET firstName=$1, lastName=$2, phone=$3, email=$4, status=$5, role=$6 
				  WHERE id=$7`

		result, err := db.ExecContext(ctx, query,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.Email,
			user.Status,
			user.Role,
			userID,
		)

		if err != nil {
			log.Printf("UpdateUser error: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Database error",
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Database error",
			})
			return
		}

		if rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, Response{
				Success: false,
				Message: "User not found",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Success: true,
			Message: "User updated successfully",
		})
	}
}

// DELETE /users/{id} - Soft delete user
func DeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID, err := parseIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid or missing User ID",
			})
			return
		}

		// Soft delete directly & check affected rows (Optimized 1 DB Call)
		query := `UPDATE users SET status = 0 WHERE id = $1`
		result, err := db.ExecContext(ctx, query, userID)

		if err != nil {
			log.Printf("DeleteUser error: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Database error",
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, Response{
				Success: false,
				Message: "User not found",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Success: true,
			Message: "User deactivated successfully",
		})
	}
}
