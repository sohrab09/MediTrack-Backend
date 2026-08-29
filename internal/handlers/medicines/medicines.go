package medicines

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"meditrack-backend/internal/models"
)

// Response standard structure
type Response struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func respondSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Status:  status,
		Success: true,
		Message: message,
		Data:    data,
	})
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Status:  status,
		Success: false,
		Message: message,
	})
}

// Get User ID from JWT Context Key
func getUserIDFromContext(ctx context.Context) *int {
	if val := ctx.Value("userID"); val != nil {
		if id, ok := val.(int); ok && id > 0 {
			return &id
		}
	}
	return nil
}

// AddMedicine - Create a new medicine entry
func AddMedicine(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Medicine
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		if m.Name == "" || m.Code == "" || m.CategoryID == 0 || m.BoxSizeID == 0 || m.UnitID == 0 {
			respondError(w, http.StatusBadRequest, "Required fields (code, name, category_id, box_size_id, unit_id) are missing")
			return
		}

		// JWT Context থেকে User ID নেওয়া
		m.CreatedBy = getUserIDFromContext(r.Context())

		if m.Status == 0 {
			m.Status = 1
		}

		query := `
			INSERT INTO medicines (
				code, barcode, name, strength, generic, category_id, type_id, 
				box_size_id, unit_id, leaf_id, selling_price, mrp, current_stock, minimum_stock, status, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			RETURNING id, created_at, updated_at
		`

		err := db.QueryRowContext(
			r.Context(), query,
			m.Code, m.Barcode, m.Name, m.Strength, m.Generic, m.CategoryID, m.TypeID,
			m.BoxSizeID, m.UnitID, m.LeafID, m.SellingPrice, m.MRP, m.CurrentStock, m.MinimumStock, m.Status, m.CreatedBy,
		).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to insert medicine: "+err.Error())
			return
		}

		respondSuccess(w, http.StatusCreated, "Medicine added successfully", m)
	}
}

// GetMedicines - Retrieve medicine list
func GetMedicines(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
			SELECT id, code, barcode, name, strength, generic, category_id, type_id, 
			       box_size_id, unit_id, leaf_id, selling_price, mrp, current_stock, minimum_stock, status, created_by, created_at, updated_at 
			FROM medicines ORDER BY id DESC
		`

		rows, err := db.QueryContext(r.Context(), query)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Database error: "+err.Error())
			return
		}
		defer rows.Close()

		var list []models.Medicine
		for rows.Next() {
			var m models.Medicine
			if err := rows.Scan(
				&m.ID, &m.Code, &m.Barcode, &m.Name, &m.Strength, &m.Generic,
				&m.CategoryID, &m.TypeID, &m.BoxSizeID, &m.UnitID, &m.LeafID,
				&m.SellingPrice, &m.MRP, &m.CurrentStock, &m.MinimumStock,
				&m.Status, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
			); err != nil {
				respondError(w, http.StatusInternalServerError, "Error scanning row: "+err.Error())
				return
			}
			list = append(list, m)
		}
		if err := rows.Err(); err != nil {
			respondError(w, http.StatusInternalServerError, "Error iterating medicine rows: "+err.Error())
			return
		}

		respondSuccess(w, http.StatusOK, "Medicines fetched successfully", list)
	}
}

// GetMedicineByID - Retrieve single medicine details
func GetMedicineByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			respondError(w, http.StatusBadRequest, "Invalid or missing medicine ID")
			return
		}

		query := `
			SELECT id, code, barcode, name, strength, generic, category_id, type_id, 
			       box_size_id, unit_id, leaf_id, selling_price, mrp, current_stock, minimum_stock, status, created_by, created_at, updated_at 
			FROM medicines WHERE id = $1
		`

		var m models.Medicine
		err = db.QueryRowContext(r.Context(), query, id).Scan(
			&m.ID, &m.Code, &m.Barcode, &m.Name, &m.Strength, &m.Generic,
			&m.CategoryID, &m.TypeID, &m.BoxSizeID, &m.UnitID, &m.LeafID,
			&m.SellingPrice, &m.MRP, &m.CurrentStock, &m.MinimumStock,
			&m.Status, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Medicine not found")
			return
		} else if err != nil {
			respondError(w, http.StatusInternalServerError, "Database error: "+err.Error())
			return
		}

		respondSuccess(w, http.StatusOK, "Medicine details fetched", m)
	}
}

// UpdateMedicine - Update existing medicine information
func UpdateMedicine(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			respondError(w, http.StatusBadRequest, "Invalid or missing medicine ID")
			return
		}

		var m models.Medicine
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		query := `
			UPDATE medicines 
			SET code = $1, barcode = $2, name = $3, strength = $4, generic = $5, 
			    category_id = $6, type_id = $7, box_size_id = $8, unit_id = $9, leaf_id = $10, 
			    selling_price = $11, mrp = $12, minimum_stock = $13, status = $14, updated_at = CURRENT_TIMESTAMP 
			WHERE id = $15
			RETURNING updated_at
		`

		err = db.QueryRowContext(
			r.Context(), query,
			m.Code, m.Barcode, m.Name, m.Strength, m.Generic,
			m.CategoryID, m.TypeID, m.BoxSizeID, m.UnitID, m.LeafID,
			m.SellingPrice, m.MRP, m.MinimumStock, m.Status, id,
		).Scan(&m.UpdatedAt)

		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Medicine not found")
			return
		} else if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update medicine: "+err.Error())
			return
		}

		m.ID = id
		respondSuccess(w, http.StatusOK, "Medicine updated successfully", m)
	}
}

// DeleteMedicine - Remove a medicine record
func DeleteMedicine(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			respondError(w, http.StatusBadRequest, "Invalid or missing medicine ID")
			return
		}

		result, err := db.ExecContext(r.Context(), "DELETE FROM medicines WHERE id = $1", id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to delete medicine: "+err.Error())
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			respondError(w, http.StatusNotFound, "Medicine not found")
			return
		}

		respondSuccess(w, http.StatusOK, "Medicine deleted successfully", nil)
	}
}
