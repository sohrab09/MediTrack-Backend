package medicine_leaves

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
func validateInput(data *models.MedicineLeaf) error {
	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" {
		return errors.New("leaf name is required")
	}
	if data.QtyPerLeaf <= 0 {
		return errors.New("qty_per_leaf must be greater than 0")
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

// CREATE - Create a new medicine leaf
func CreateMedicineLeaf(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var leaf models.MedicineLeaf
		if err := json.NewDecoder(r.Body).Decode(&leaf); err != nil {
			respondJSON(w, http.StatusBadRequest, "Invalid JSON body format", false, nil)
			return
		}

		// Default active setting if not provided
		if leaf.Status == 0 {
			leaf.Status = 1
		}

		if err := validateInput(&leaf); err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `
			INSERT INTO medicine_leaves (name, qty_per_leaf, description, status, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, NOW(), NOW()) 
			RETURNING id, name, qty_per_leaf, COALESCE(description, ''), status, created_at, updated_at
		`

		var createdLeaf models.MedicineLeaf
		err := db.QueryRowContext(ctx, query, leaf.Name, leaf.QtyPerLeaf, leaf.Description, leaf.Status).
			Scan(&createdLeaf.ID, &createdLeaf.Name, &createdLeaf.QtyPerLeaf, &createdLeaf.Description, &createdLeaf.Status, &createdLeaf.CreatedAt, &createdLeaf.UpdatedAt)

		if err != nil {
			log.Printf("Error creating medicine leaf: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to create medicine leaf", false, nil)
			return
		}

		respondJSON(w, http.StatusCreated, "Medicine leaf created successfully", true, createdLeaf)
	}
}

// READ - Get all medicine leaves
func GetMedicineLeaves(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := `SELECT id, name, qty_per_leaf, COALESCE(description, ''), status, created_at, updated_at FROM medicine_leaves ORDER BY created_at DESC`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("Error fetching medicine leaves: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to fetch medicine leaves", false, nil)
			return
		}
		defer rows.Close()

		leaves := make([]models.MedicineLeaf, 0)

		for rows.Next() {
			var leaf models.MedicineLeaf
			if err := rows.Scan(&leaf.ID, &leaf.Name, &leaf.QtyPerLeaf, &leaf.Description, &leaf.Status, &leaf.CreatedAt, &leaf.UpdatedAt); err != nil {
				log.Printf("Error scanning medicine leaf: %v", err)
				respondJSON(w, http.StatusInternalServerError, "Error parsing database records", false, nil)
				return
			}
			leaves = append(leaves, leaf)
		}

		if err = rows.Err(); err != nil {
			log.Printf("Error iterating rows: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Error reading rows", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine leaves fetched successfully", true, leaves)
	}
}

// READ - Get single medicine leaf by ID
func GetMedicineLeafByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `SELECT id, name, qty_per_leaf, COALESCE(description, ''), status, created_at, updated_at FROM medicine_leaves WHERE id = $1`

		var leaf models.MedicineLeaf
		err = db.QueryRowContext(ctx, query, id).Scan(&leaf.ID, &leaf.Name, &leaf.QtyPerLeaf, &leaf.Description, &leaf.Status, &leaf.CreatedAt, &leaf.UpdatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, "Medicine leaf not found", false, nil)
			return
		}

		if err != nil {
			log.Printf("Error fetching medicine leaf: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to fetch medicine leaf", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine leaf fetched successfully", true, leaf)
	}
}

// UPDATE - Update a medicine leaf
func UpdateMedicineLeaf(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		var leaf models.MedicineLeaf

		if err := json.NewDecoder(r.Body).Decode(&leaf); err != nil {
			respondJSON(w, http.StatusBadRequest, "Invalid JSON body format", false, nil)
			return
		}

		if err := validateInput(&leaf); err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `
			UPDATE medicine_leaves 
			SET name = $1, qty_per_leaf = $2, description = $3, status = $4, updated_at = NOW() 
			WHERE id = $5 
			RETURNING id, name, qty_per_leaf, COALESCE(description, ''), status, created_at, updated_at
		`

		var updatedLeaf models.MedicineLeaf
		err = db.QueryRowContext(ctx, query, leaf.Name, leaf.QtyPerLeaf, leaf.Description, leaf.Status, id).
			Scan(&updatedLeaf.ID, &updatedLeaf.Name, &updatedLeaf.QtyPerLeaf, &updatedLeaf.Description, &updatedLeaf.Status, &updatedLeaf.CreatedAt, &updatedLeaf.UpdatedAt)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, "Medicine leaf not found", false, nil)
			return
		}

		if err != nil {
			log.Printf("Error updating medicine leaf: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to update medicine leaf", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine leaf updated successfully", true, updatedLeaf)
	}
}

// DELETE - Soft delete a medicine leaf (Status = 0)
func DeleteMedicineLeaf(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := getIDFromPath(r)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, err.Error(), false, nil)
			return
		}

		query := `UPDATE medicine_leaves SET status = 0, updated_at = NOW() WHERE id = $1`

		result, err := db.ExecContext(ctx, query, id)
		if err != nil {
			log.Printf("Error soft deleting medicine leaf: %v", err)
			respondJSON(w, http.StatusInternalServerError, "Failed to delete medicine leaf", false, nil)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, "Medicine leaf not found or already deleted", false, nil)
			return
		}

		respondJSON(w, http.StatusOK, "Medicine leaf deleted successfully", true, nil)
	}
}
