package getmedicinecategories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"meditrack-backend/internal/models"
	"net/http"
	"strconv"
	"time"
)

// Response structure
type Response struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON response helper
func respondJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

func getIDFromPath(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return 0, errors.New("id parameter is required")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("invalid id format")
	}
	return id, nil
}

func GetMedicineCategories(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			respondJSON(w, http.StatusMethodNotAllowed, Response{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		query := `SELECT id, name, status, created_at FROM categories`

		rows, err := db.Query(query)
		if err != nil {
			log.Println("Query error:", err) // 🔥 important
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Database error",
			})
			return
		}
		defer rows.Close()

		// 🔥 Response struct (formatted time)
		type CategoryResponse struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Status    int    `json:"status"`
			CreatedAt string `json:"created_at"`
		}

		var categories []CategoryResponse

		for rows.Next() {
			var category models.MedicineCategories

			if err := rows.Scan(
				&category.ID,
				&category.Name,
				&category.Status,
				&category.CreatedAt,
			); err != nil {

				log.Println("Scan error:", err) // 🔥 critical
				respondJSON(w, http.StatusInternalServerError, Response{
					Success: false,
					Message: "Database error",
				})
				return
			}

			// 🔥 Convert time → string for frontend
			categories = append(categories, CategoryResponse{
				ID:        category.ID,
				Name:      category.Name,
				Status:    category.Status,
				CreatedAt: category.CreatedAt.Format(time.RFC3339),
			})
		}

		// 🔥 check iteration error
		if err := rows.Err(); err != nil {
			log.Println("Rows error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Database error",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Success: true,
			Message: "Categories fetched successfully",
			Data:    categories,
		})
	}
}

func GetMedicineCategoryByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		query := `SELECT id, name, status, created_at FROM categories WHERE id = $1`

		var category models.MedicineCategories
		err = db.QueryRow(query, id).Scan(&category.ID, &category.Name, &category.Status, &category.CreatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Medicine category not found",
			})
			return
		}

		if err != nil {
			log.Printf("Error fetching medicine category: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Database error",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Category fetched successfully",
			Data:    category,
		})
	}
}
