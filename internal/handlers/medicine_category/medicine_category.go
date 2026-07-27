package medicine_category

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"meditrack-backend/internal/models"
)

// Standard API Response structure
type Response struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Global Category Response for API output consistency
type CategoryResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

type UpdateRequest struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

// Helper: Response JSON
func respondJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(res)
}

// Helper: Extract ID from URL path
func getIDFromPath(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return 0, errors.New("id parameter is required")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id format")
	}
	return id, nil
}

// Helper: Input Validation
func validateCategoryInput(name string, status int) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if status != 0 && status != 1 {
		return errors.New("status must be 0 (inactive) or 1 (active)")
	}
	return nil
}

// Helper: Convert Database Row into Clean CategoryResponse
func formatCategoryResponse(id int, name string, status int, createdAt time.Time) CategoryResponse {
	return CategoryResponse{
		ID:        id,
		Name:      name,
		Status:    status,
		CreatedAt: createdAt.Format(time.RFC3339),
	}
}

// ---------------------------------------------------------
// Handlers
// ---------------------------------------------------------

// CREATE
func CreateMedicineCategories(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req struct {
			Data models.MedicineCategories `json:"data"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid JSON format",
			})
			return
		}

		data := req.Data

		if err := validateCategoryInput(data.Name, data.Status); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// Duplicate Check with Context
		var existingID int
		err := db.QueryRowContext(ctx, "SELECT id FROM medicine_categories WHERE LOWER(name) = LOWER($1)", data.Name).Scan(&existingID)
		if err == nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Category already exists",
			})
			return
		} else if err != sql.ErrNoRows {
			log.Println("Duplicate check error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Database query error",
			})
			return
		}

		// Insert
		query := `
			INSERT INTO medicine_categories (name, status, created_at)
			VALUES ($1, $2, $3)
			RETURNING id, created_at
		`

		var createdID int
		var createdAt time.Time

		err = db.QueryRowContext(ctx, query, data.Name, data.Status, time.Now()).Scan(&createdID, &createdAt)
		if err != nil {
			log.Println("Insert error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to create category",
			})
			return
		}

		respondJSON(w, http.StatusCreated, Response{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Category created successfully",
			Data:    formatCategoryResponse(createdID, data.Name, data.Status, createdAt),
		})
	}
}

// GET ALL
func GetMedicineCategories(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := `SELECT id, name, status, created_at FROM medicine_categories ORDER BY id DESC`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Println("Query error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Database error",
			})
			return
		}
		defer rows.Close()

		categories := make([]CategoryResponse, 0)

		for rows.Next() {
			var id, status int
			var name string
			var createdAt time.Time

			if err := rows.Scan(&id, &name, &status, &createdAt); err != nil {
				log.Println("Scan error:", err)
				respondJSON(w, http.StatusInternalServerError, Response{
					Status:  http.StatusInternalServerError,
					Success: false,
					Message: "Database error",
				})
				return
			}

			categories = append(categories, formatCategoryResponse(id, name, status, createdAt))
		}

		if err := rows.Err(); err != nil {
			log.Println("Rows iteration error:", err)
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
			Message: "Categories fetched successfully",
			Data:    categories,
		})
	}
}

// GET BY ID
func GetMedicineCategoryByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		query := `SELECT id, name, status, created_at FROM medicine_categories WHERE id = $1`

		var categoryID, status int
		var name string
		var createdAt time.Time

		err = db.QueryRowContext(ctx, query, id).Scan(&categoryID, &name, &status, &createdAt)

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
			Data:    formatCategoryResponse(categoryID, name, status, createdAt),
		})
	}
}

// UPDATE
func UpdateMedicineCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		var req UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid JSON format",
			})
			return
		}

		if err := validateCategoryInput(req.Name, req.Status); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		query := `
			UPDATE medicine_categories
			SET name = $1, status = $2
			WHERE id = $3
			RETURNING id, name, status, created_at
		`

		var updatedID, status int
		var name string
		var createdAt time.Time

		// FIXED: Scanned into time.Time instead of string to prevent runtime crash
		err = db.QueryRowContext(ctx, query, req.Name, req.Status, id).Scan(&updatedID, &name, &status, &createdAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Category not found",
			})
			return
		}

		if err != nil {
			log.Println("Update error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to update category",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Category updated successfully",
			Data:    formatCategoryResponse(updatedID, name, status, createdAt),
		})
	}
}

// DELETE
func DeleteMedicineCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		result, err := db.ExecContext(ctx, "DELETE FROM medicine_categories WHERE id = $1", id)
		if err != nil {
			log.Println("Delete error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to delete category",
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Category not found",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Category deleted successfully",
		})
	}
}
