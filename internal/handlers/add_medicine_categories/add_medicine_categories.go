package addmedicinecategories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"meditrack-backend/internal/models"
	"net/http"
	"strings"
	"time"
)

// Response structure
type Response struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON helper
func respondJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

// ✅ Separate response struct (IMPORTANT)
type CategoryResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

// Validation
func validateCategoryInput(data *models.MedicineCategories) error {
	if strings.TrimSpace(data.Name) == "" {
		return errors.New("name is required")
	}
	if data.Status != 0 && data.Status != 1 {
		return errors.New("status must be 0 (inactive) or 1 (active)")
	}
	return nil
}

func CreateMedicineCategories(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			respondJSON(w, http.StatusMethodNotAllowed, Response{
				Status:  http.StatusMethodNotAllowed,
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		ctx := r.Context()

		var req struct {
			Data models.MedicineCategories `json:"data"`
		}

		// Decode request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid JSON format",
			})
			return
		}

		data := req.Data

		// Validate
		if err := validateCategoryInput(&data); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// Duplicate check
		var existingID int
		err := db.QueryRowContext(ctx,
			"SELECT id FROM categories WHERE name = $1",
			data.Name,
		).Scan(&existingID)

		if err != nil && err != sql.ErrNoRows {
			log.Println("Duplicate check error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Database query error",
			})
			return
		}

		if err == nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Category already exists",
			})
			return
		}

		// Insert
		query := `
			INSERT INTO categories (name, status, created_at)
			VALUES ($1, $2, $3)
			RETURNING id, created_at
		`

		var createdID int
		var createdAt time.Time

		err = db.QueryRowContext(
			ctx,
			query,
			data.Name,
			data.Status,
			time.Now(),
		).Scan(&createdID, &createdAt)

		if err != nil {
			log.Println("Insert error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to create category",
			})
			return
		}

		// ✅ Build response (DON'T touch model)
		response := CategoryResponse{
			ID:        createdID,
			Name:      data.Name,
			Status:    data.Status,
			CreatedAt: createdAt.Format(time.RFC3339),
		}

		log.Printf("Created category: %+v", response)

		respondJSON(w, http.StatusCreated, Response{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Category created successfully",
			Data:    response,
		})
	}
}
