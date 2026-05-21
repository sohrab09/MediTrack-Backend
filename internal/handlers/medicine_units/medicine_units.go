package medicineunits

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"meditrack-backend/internal/models"
	"net/http"
	"strconv"
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

// Validation
func validateInput(data *models.MedicineUnits) error {
	if strings.TrimSpace(data.Name) == "" {
		return errors.New("name is required")
	}
	if data.Status != 0 && data.Status != 1 {
		return errors.New("status must be 0 (inactive) or 1 (active)")
	}
	return nil
}

// Extract ID from path parameter
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

// CREATE - Create a new medicine unit
func CreateMedicineUnits(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Data models.MedicineUnits `json:"data"`
		}

		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		data := req.Data

		if err := validateInput(&data); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// Insert into database
		query := `INSERT INTO medicine_units (name, status, created_at) 
				 VALUES ($1, $2, $3) RETURNING id, created_at`

		var createdUnit models.MedicineUnits
		err := db.QueryRow(query, data.Name, data.Status, time.Now()).
			Scan(&createdUnit.ID, &createdUnit.CreatedAt)

		if err != nil {
			log.Printf("Error creating medicine unit: %v", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to create medicine unit",
			})
			return
		}

		createdUnit.Name = data.Name
		createdUnit.Status = data.Status

		respondJSON(w, http.StatusCreated, Response{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Medicine unit created successfully",
			Data:    createdUnit,
		})
	}
}

// READ - Get all medicine units
func GetMedicineUnits(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `SELECT id, name, status, created_at FROM medicine_units ORDER BY created_at DESC`

		rows, err := db.Query(query)
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

		var units []models.MedicineUnits
		for rows.Next() {
			var unit models.MedicineUnits
			if err := rows.Scan(&unit.ID, &unit.Name, &unit.Status, &unit.CreatedAt); err != nil {
				log.Printf("Error scanning medicine unit: %v", err)
				respondJSON(w, http.StatusInternalServerError, Response{
					Status:  http.StatusInternalServerError,
					Success: false,
					Message: "Error parsing data",
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
				Message: "Error fetching data",
			})
			return
		}

		if units == nil {
			units = []models.MedicineUnits{}
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
		err = db.QueryRow(query, id).Scan(&unit.ID, &unit.Name, &unit.Status, &unit.CreatedAt)

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
		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		var req struct {
			Data models.MedicineUnits `json:"data"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		data := req.Data

		if err := validateInput(&data); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// Check if record exists
		var exists bool
		err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM medicine_units WHERE id = $1)`, id).Scan(&exists)
		if err != nil || !exists {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Medicine unit not found",
			})
			return
		}

		// Update record
		query := `UPDATE medicine_units SET name = $1, status = $2 WHERE id = $3 
				 RETURNING id, name, status, created_at`

		var updatedUnit models.MedicineUnits
		err = db.QueryRow(query, data.Name, data.Status, id).
			Scan(&updatedUnit.ID, &updatedUnit.Name, &updatedUnit.Status, &updatedUnit.CreatedAt)

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

// DELETE - delete a medicine unit
func DeleteMedicineUnits(db *sql.DB) http.HandlerFunc {
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

		// Check if record exists
		var exists bool
		err = db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM medicine_units WHERE id = $1)`,
			id,
		).Scan(&exists)

		if err != nil {
			log.Printf("Error checking medicine unit: %v", err)

			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Database error",
			})
			return
		}

		if !exists {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Medicine unit not found",
			})
			return
		}

		// Soft delete (status = 0)
		query := `UPDATE medicine_units SET status = 0 WHERE id = $1`

		_, err = db.Exec(query, id)

		if err != nil {
			log.Printf("Error soft deleting medicine unit: %v", err)

			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to delete medicine unit",
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
