package medicine_box_sizes

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"meditrack-backend/internal/models"
)

// Standard API Response structure
type Response struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON Helper
func respondJSON(w http.ResponseWriter, status int, message string, success bool, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{
		Status:  status,
		Success: success,
		Message: message,
		Data:    data,
	})
}

// Validation Helper
func validateInput(data *models.MedicineBoxSize) error {
	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" {
		return errors.New("box size name is required")
	}
	if data.TotalPcs <= 0 {
		return errors.New("total_pcs must be greater than 0")
	}
	if data.Status != 0 && data.Status != 1 {
		return errors.New("status must be 0 (inactive) or 1 (active)")
	}
	return nil
}

// Extract ID from URL Path parameter
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

// CREATE - Create a new medicine box size
func CreateMedicineBoxSize(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var boxSize models.MedicineBoxSize
		if err := json.NewDecoder(r.Body).Decode(&boxSize); err != nil {
			respondJSON(w, http.StatusBadRequest, "Invalid JSON body format", false, nil)
			return
		}

		// Default active setting if not provided
		if boxSize.Status == 0 {
			boxSize.Status = 1
		}

		if err := validateInput(&boxSize); err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `
			INSERT INTO medicine_box_sizes (name, total_pcs, status, created_at, updated_at) 
			VALUES ($1, $2, $3, NOW(), NOW()) 
			RETURNING id, name, total_pcs, status, created_at, updated_at
		`

		var createdBoxSize models.MedicineBoxSize
		err := db.QueryRowContext(ctx, query, boxSize.Name, boxSize.TotalPcs, boxSize.Status).
			Scan(&createdBoxSize.ID, &createdBoxSize.Name, &createdBoxSize.TotalPcs, &createdBoxSize.Status, &createdBoxSize.CreatedAt, &createdBoxSize.UpdatedAt)

		if err != nil {
			log.Printf("Error creating medicine box size: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to create medicine box size or size already exists", false, nil)
			return
		}

		respondJSON(w, http.StatusCreated, "Medicine box size created successfully", true, createdBoxSize)
	}
}

// READ - Get all medicine box sizes
func GetMedicineBoxSizes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := `SELECT id, name, total_pcs, status, created_at, updated_at FROM medicine_box_sizes ORDER BY created_at DESC`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("Error fetching medicine box sizes: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to fetch medicine box sizes", false, nil)
			return
		}
		defer rows.Close()

		boxSizes := make([]models.MedicineBoxSize, 0)

		for rows.Next() {
			var boxSize models.MedicineBoxSize
			if err := rows.Scan(&boxSize.ID, &boxSize.Name, &boxSize.TotalPcs, &boxSize.Status, &boxSize.CreatedAt, &boxSize.UpdatedAt); err != nil {
				log.Printf("Error scanning medicine box size: %v", err)
				respondJSON(w, http.StatusInternalServerError, "Error parsing database records", false, nil)
				return
			}
			boxSizes = append(boxSizes, boxSize)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Error iterating rows: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Error reading rows", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine box sizes fetched successfully", true, boxSizes)
	}
}

// READ - Get single medicine box size by ID
func GetMedicineBoxSizeByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `SELECT id, name, total_pcs, status, created_at, updated_at FROM medicine_box_sizes WHERE id = $1`

		var boxSize models.MedicineBoxSize
		err = db.QueryRowContext(ctx, query, id).Scan(&boxSize.ID, &boxSize.Name, &boxSize.TotalPcs, &boxSize.Status, &boxSize.CreatedAt, &boxSize.UpdatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, "Medicine box size not found", false, nil)
			return
		}

		if err != nil {
			log.Printf("Error fetching medicine box size: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to fetch medicine box size", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine box size fetched successfully", true, boxSize)
	}
}

// UPDATE - Update a medicine box size
func UpdateMedicineBoxSize(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		var boxSize models.MedicineBoxSize

		if err := json.NewDecoder(r.Body).Decode(&boxSize); err != nil {
			respondJSON(w, http.StatusBadRequest, "Invalid JSON body format", false, nil)
			return
		}

		if err := validateInput(&boxSize); err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `
			UPDATE medicine_box_sizes 
			SET name = $1, total_pcs = $2, status = $3, updated_at = NOW() 
			WHERE id = $4 
			RETURNING id, name, total_pcs, status, created_at, updated_at
		`

		var updatedBoxSize models.MedicineBoxSize
		err = db.QueryRowContext(ctx, query, boxSize.Name, boxSize.TotalPcs, boxSize.Status, id).
			Scan(&updatedBoxSize.ID, &updatedBoxSize.Name, &updatedBoxSize.TotalPcs, &updatedBoxSize.Status, &updatedBoxSize.CreatedAt, &updatedBoxSize.UpdatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, "Medicine box size not found", false, nil)
			return
		}

		if err != nil {
			log.Printf("Error updating medicine box size: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to update medicine box size", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine box size updated successfully", true, updatedBoxSize)
	}
}

// DELETE - Soft delete a medicine box size (Status = 0)
func DeleteMedicineBoxSize(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `UPDATE medicine_box_sizes SET status = 0, updated_at = NOW() WHERE id = $1`

		result, err := db.ExecContext(ctx, query, id)
		if err != nil {
			log.Printf("Error soft deleting medicine box size: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to delete medicine box size", false, nil)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, "Medicine box size not found or already deleted", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine box size deleted successfully", true, nil)
	}
}
