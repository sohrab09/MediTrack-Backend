package medicine_types

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

// JSON Helper (স্বয়ংক্রিয়ভাবে HTTP Status কোড যুক্ত করবে)
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
func validateInput(data *models.MedicineType) error {
	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" {
		return errors.New("name is required")
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

// CREATE - Create a new medicine type
func CreateMedicineTypes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var medType models.MedicineType

		if err := json.NewDecoder(r.Body).Decode(&medType); err != nil {
			respondJSON(w, http.StatusBadRequest, "Invalid JSON body format", false, nil)
			return
		}

		if err := validateInput(&medType); err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		// টেবিল নাম medicine_types হিসেবে সংশোধিত
		query := `
			INSERT INTO medicine_types (name, status, created_at) 
			VALUES ($1, $2, NOW()) 
			RETURNING id, name, status, created_at
		`

		var createdType models.MedicineType
		err := db.QueryRowContext(ctx, query, medType.Name, medType.Status).
			Scan(&createdType.ID, &createdType.Name, &createdType.Status, &createdType.CreatedAt)

		if err != nil {
			log.Printf("Error creating medicine type: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to create medicine type", false, nil)
			return
		}

		respondJSON(w, http.StatusCreated, "Medicine type created successfully", true, createdType)
	}
}

// READ - Get all medicine types
func GetMedicineTypes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := `SELECT id, name, status, created_at FROM medicine_types ORDER BY created_at DESC`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("Error fetching medicine types: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to fetch medicine types", false, nil)
			return
		}
		defer rows.Close()

		types := make([]models.MedicineType, 0)

		for rows.Next() {
			var medType models.MedicineType
			if err := rows.Scan(&medType.ID, &medType.Name, &medType.Status, &medType.CreatedAt); err != nil {
				log.Printf("Error scanning medicine type: %v", err)
				respondJSON(w, http.StatusInternalServerError, "Error parsing database records", false, nil)
				return
			}
			types = append(types, medType)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Error iterating rows: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Error reading rows", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine types fetched successfully", true, types)
	}
}

// READ - Get single medicine type by ID
func GetMedicineTypesByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `SELECT id, name, status, created_at FROM medicine_types WHERE id = $1`

		var medType models.MedicineType
		err = db.QueryRowContext(ctx, query, id).Scan(&medType.ID, &medType.Name, &medType.Status, &medType.CreatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, "Medicine type not found", false, nil)
			return
		}

		if err != nil {
			log.Printf("Error fetching medicine type: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to fetch medicine type", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine type fetched successfully", true, medType)
	}
}

// UPDATE - Update a medicine type
func UpdateMedicineTypes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		var medType models.MedicineType

		if err := json.NewDecoder(r.Body).Decode(&medType); err != nil {
			respondJSON(w, http.StatusBadRequest, "Invalid JSON body format", false, nil)
			return
		}

		if err := validateInput(&medType); err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `
			UPDATE medicine_types 
			SET name = $1, status = $2 
			WHERE id = $3 
			RETURNING id, name, status, created_at
		`

		var updatedType models.MedicineType
		err = db.QueryRowContext(ctx, query, medType.Name, medType.Status, id).
			Scan(&updatedType.ID, &updatedType.Name, &updatedType.Status, &updatedType.CreatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, "Medicine type not found", false, nil)
			return
		}

		if err != nil {
			log.Printf("Error updating medicine type: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to update medicine type", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine type updated successfully", true, updatedType)
	}
}

// DELETE - Soft delete a medicine type
func DeleteMedicineTypes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `UPDATE medicine_types SET status = 0 WHERE id = $1`

		result, err := db.ExecContext(ctx, query, id)
		if err != nil {
			log.Printf("Error soft deleting medicine type: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to delete medicine type", false, nil)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, "Medicine type not found or already deleted", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine type deleted successfully", true, nil)
	}
}
