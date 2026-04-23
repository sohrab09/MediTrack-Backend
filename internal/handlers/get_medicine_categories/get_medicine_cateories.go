package getmedicinecategories

import (
	"database/sql"
	"encoding/json"
	"log"
	"meditrack-backend/internal/models"
	"net/http"
	"time"
)

// Response structure
type Response struct {
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
