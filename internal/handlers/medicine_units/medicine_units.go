package medicine_units

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

// Global Standard API Response structure
type Response struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON Helper
func respondJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(res)
}

// Validation Helper
func validateInput(data *models.MedicineUnits) error {
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

// CREATE - Create a new medicine unit
func CreateMedicineUnits(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var unit models.MedicineUnits

		// Directly decode JSON without wrapper
		if err := json.NewDecoder(r.Body).Decode(&unit); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid JSON body format",
			})
			return
		}

		if err := validateInput(&unit); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// Insert into DB (Letting DB assign created_at timestamp automatically)
		query := `
			INSERT INTO medicine_units (name, status, created_at) 
			VALUES ($1, $2, NOW()) 
			RETURNING id, name, status, created_at
		`

		var createdUnit models.MedicineUnits
		err := db.QueryRowContext(ctx, query, unit.Name, unit.Status).
			Scan(&createdUnit.ID, &createdUnit.Name, &createdUnit.Status, &createdUnit.CreatedAt)

		if err != nil {
			log.Printf("Error creating medicine unit: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to create medicine unit",
			})
			return
		}

		respondJSON(w, http.StatusCreated, Response{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Medicine unit created successfully",
			Data:    createdUnit,
		})
	}
}

// READ - Get all active/all medicine units
func GetMedicineUnits(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := `SELECT id, name, status, created_at FROM medicine_units ORDER BY created_at DESC`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("Error fetching medicine units: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to fetch medicine units",
			})
			return
		}
		defer rows.Close()

		units := make([]models.MedicineUnits, 0)

		for rows.Next() {
			var unit models.MedicineUnits
			if err := rows.Scan(&unit.ID, &unit.Name, &unit.Status, &unit.CreatedAt); err != nil {
				log.Printf("Error scanning medicine unit: %v", err)
				respondJSON(w, http.StatusInternalServerError, Response{
					Status:  http.StatusInternalServerError,
					Success: false,
					Message: "Error parsing database records",
				})
				return
			}
			units = append(units, unit)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Error iterating rows: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Error reading rows",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Medicine units fetched successfully",
			Data:    units,
		})
	}
}

// READ - Get single medicine unit by ID
func GetMedicineUnitByID(db *sql.DB) http.HandlerFunc {
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

		query := `SELECT id, name, status, created_at FROM medicine_units WHERE id = $1`

		var unit models.MedicineUnits
		err = db.QueryRowContext(ctx, query, id).Scan(&unit.ID, &unit.Name, &unit.Status, &unit.CreatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Medicine unit not found",
			})
			return
		}

		if err != nil {
			log.Printf("Error fetching medicine unit: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to fetch medicine unit",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Medicine unit fetched successfully",
			Data:    unit,
		})
	}
}

// UPDATE - Update a medicine unit
func UpdateMedicineUnits(db *sql.DB) http.HandlerFunc {
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

		var unit models.MedicineUnits

		if err := json.NewDecoder(r.Body).Decode(&unit); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid JSON body format",
			})
			return
		}

		if err := validateInput(&unit); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		query := `
			UPDATE medicine_units 
			SET name = $1, status = $2 
			WHERE id = $3 
			RETURNING id, name, status, created_at
		`

		var updatedUnit models.MedicineUnits
		err = db.QueryRowContext(ctx, query, unit.Name, unit.Status, id).
			Scan(&updatedUnit.ID, &updatedUnit.Name, &updatedUnit.Status, &updatedUnit.CreatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Medicine unit not found",
			})
			return
		}

		if err != nil {
			log.Printf("Error updating medicine unit: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to update medicine unit",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Medicine unit updated successfully",
			Data:    updatedUnit,
		})
	}
}

// DELETE - Soft delete a medicine unit
func DeleteMedicineUnits(db *sql.DB) http.HandlerFunc {
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

		query := `UPDATE medicine_units SET status = 0 WHERE id = $1`

		result, err := db.ExecContext(ctx, query, id)
		if err != nil {
			log.Printf("Error soft deleting medicine unit: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to delete medicine unit",
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Medicine unit not found or already deleted",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Medicine unit deleted successfully",
		})
	}
}
