package medicines

import (
	"database/sql"
	"encoding/json"
	"io"
	"meditrack-backend/internal/models"
	"net/http"
	"strconv"
	"strings"
)

type Response struct {
	Status  bool        `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, resp Response) {
	resp.Code = status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func parseIDFromURL(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	return strconv.Atoi(idStr)
}

// -------------------------------------------------------------
// 1. ADD MEDICINE (POST /api/v1/medicines)
// -------------------------------------------------------------
func AddMedicineHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, Response{
				Status:  false,
				Message: "Method not allowed",
				Error:   "Only POST method is allowed for this endpoint",
			})
			return
		}

		var m models.Medicine

		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Invalid Request Payload",
				Error:   err.Error(),
			})
			return
		}

		// ----------------------------
		// Validation
		// ----------------------------
		if strings.TrimSpace(m.Name) == "" {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Medicine name is required",
			})
			return
		}

		if m.CategoryID <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Category is required",
			})
			return
		}

		if m.BoxSizeID <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Box Size is required",
			})
			return
		}

		if m.UnitID <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Unit is required",
			})
			return
		}

		if m.MRP <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "MRP must be greater than zero",
			})
			return
		}

		if m.SellingPrice <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Selling Price must be greater than zero",
			})
			return
		}

		if m.SellingPrice > m.MRP {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Selling Price cannot be greater than MRP",
			})
			return
		}

		if m.MinimumStock < 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Minimum Stock cannot be negative",
			})
			return
		}

		// ----------------------------
		// Default Values
		// ----------------------------
		if m.Status == 0 {
			m.Status = 1
		}

		m.CurrentStock = 0

		// ----------------------------
		// Duplicate Check
		// ----------------------------
		var existingID int

		err := db.QueryRow(
			`
			SELECT id
			FROM medicines
			WHERE LOWER(name)=LOWER($1)
			AND COALESCE(strength,'')=COALESCE($2,'')
			AND status=1
			LIMIT 1
			`,
			m.Name,
			m.Strength,
		).Scan(&existingID)

		if err == nil {
			writeJSON(w, http.StatusConflict, Response{
				Status:  false,
				Message: "Duplicate Medicine",
				Error:   "Medicine already exists",
			})
			return
		}

		if err != sql.ErrNoRows {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Database Error",
				Error:   err.Error(),
			})
			return
		}

		// ----------------------------
		// Generate Medicine Code
		// (Temporary)
		// ----------------------------
		err = db.QueryRow(`
			SELECT
			'MED' || LPAD((COALESCE(MAX(id),0)+1)::TEXT,6,'0')
			FROM medicines
		`).Scan(&m.Code)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Database Error",
				Error:   "Failed to generate medicine code",
			})
			return
		}

		// ----------------------------
		// Insert
		// ----------------------------
		query := `
		INSERT INTO medicines
		(
			code,
			barcode,
			name,
			strength,
			generic,
			category_id,
			type_id,
			box_size_id,
			unit_id,
			leaf_id,
			selling_price,
			mrp,
			current_stock,
			minimum_stock,
			status
		)
		VALUES
		(
			$1,$2,$3,$4,$5,
			$6,$7,$8,$9,$10,
			$11,$12,$13,$14,$15
		)
		RETURNING
			id,
			created_at,
			updated_at
		`

		err = db.QueryRow(
			query,
			m.Code,
			m.Barcode,
			m.Name,
			m.Strength,
			m.Generic,
			m.CategoryID,
			m.TypeID,
			m.BoxSizeID,
			m.UnitID,
			m.LeafID,
			m.SellingPrice,
			m.MRP,
			m.CurrentStock,
			m.MinimumStock,
			m.Status,
		).Scan(
			&m.ID,
			&m.CreatedAt,
			&m.UpdatedAt,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Database Error",
				Error:   "Failed to insert medicine: " + err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusCreated, Response{
			Status:  true,
			Message: "Medicine added successfully",
			Data:    m,
		})
	}
}

// -------------------------------------------------------------
// 2. GET ALL MEDICINES (GET /api/v1/medicines)
// -------------------------------------------------------------
func GetMedicinesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, Response{
				Status:  false,
				Message: "Method not allowed",
				Error:   "Only GET method is allowed for this endpoint",
			})
			return
		}

		query := `
			SELECT
				id,
				code,
				barcode,
				name,
				strength,
				generic,
				category_id,
				type_id,
				box_size_id,
				unit_id,
				leaf_id,
				selling_price,
				mrp,
				current_stock,
				minimum_stock,
				status,
				created_at,
				updated_at
			FROM medicines
			WHERE status IN (0,1)
			ORDER BY id DESC
		`

		rows, err := db.Query(query)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Database Error",
				Error:   "Database query error: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		var medicinesList []models.Medicine

		for rows.Next() {
			var m models.Medicine

			err := rows.Scan(
				&m.ID,
				&m.Code,
				&m.Barcode,
				&m.Name,
				&m.Strength,
				&m.Generic,
				&m.CategoryID,
				&m.TypeID,
				&m.BoxSizeID,
				&m.UnitID,
				&m.LeafID,
				&m.SellingPrice,
				&m.MRP,
				&m.CurrentStock,
				&m.MinimumStock,
				&m.Status,
				&m.CreatedAt,
				&m.UpdatedAt,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, Response{
					Status:  false,
					Message: "Data Processing Error",
					Error:   "Error scanning row: " + err.Error(),
				})
				return
			}

			medicinesList = append(medicinesList, m)
		}

		if err = rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Data Processing Error",
				Error:   "Error after iterating rows: " + err.Error(),
			})
			return
		}

		if medicinesList == nil {
			medicinesList = []models.Medicine{}
		}

		writeJSON(w, http.StatusOK, Response{
			Status:  true,
			Message: "Medicines retrieved successfully",
			Data:    medicinesList,
		})
	}
}

// -------------------------------------------------------------
// 3. GET MEDICINE BY ID (GET /api/v1/medicines/{id})
// -------------------------------------------------------------
func GetMedicineByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, Response{
				Status:  false,
				Message: "Method not allowed",
				Error:   "Only GET method is allowed for this endpoint",
			})
			return
		}

		id, err := parseIDFromURL(r)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Invalid medicine ID",
			})
			return
		}

		query := `
			SELECT
				id,
				code,
				barcode,
				name,
				strength,
				generic,
				category_id,
				type_id,
				box_size_id,
				unit_id,
				leaf_id,
				selling_price,
				mrp,
				current_stock,
				minimum_stock,
				status,
				created_at,
				updated_at
			FROM medicines
			WHERE id = $1 AND status = 1
		`

		var m models.Medicine

		err = db.QueryRow(query, id).Scan(
			&m.ID,
			&m.Code,
			&m.Barcode,
			&m.Name,
			&m.Strength,
			&m.Generic,
			&m.CategoryID,
			&m.TypeID,
			&m.BoxSizeID,
			&m.UnitID,
			&m.LeafID,
			&m.SellingPrice,
			&m.MRP,
			&m.CurrentStock,
			&m.MinimumStock,
			&m.Status,
			&m.CreatedAt,
			&m.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, Response{
				Status:  false,
				Message: "Not Found",
				Error:   "Medicine not found",
			})
			return
		}

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Database Error",
				Error:   "Database error: " + err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, Response{
			Status:  true,
			Message: "Medicine retrieved successfully",
			Data:    m,
		})
	}
}

// -------------------------------------------------------------
// 4. UPDATE MEDICINE (PUT /api/v1/medicines/{id})
// -------------------------------------------------------------
func UpdateMedicineHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			writeJSON(w, http.StatusMethodNotAllowed, Response{
				Status:  false,
				Message: "Method not allowed",
				Error:   "Only PUT method is allowed for this endpoint",
			})
			return
		}

		id, err := parseIDFromURL(r)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Invalid medicine ID",
			})
			return
		}

		var m models.Medicine
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			errMsg := err.Error()
			if err == io.EOF {
				errMsg = "Request body is empty. Please provide medicine data in JSON format."
			}

			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Invalid Request Payload",
				Error:   errMsg,
			})
			return
		}

		// Validation
		if strings.TrimSpace(m.Name) == "" {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Medicine Name is required",
			})
			return
		}

		if m.CategoryID <= 0 || m.BoxSizeID <= 0 || m.UnitID <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Category, Box Size and Unit are required",
			})
			return
		}

		if m.SellingPrice < 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Selling Price cannot be negative",
			})
			return
		}

		if m.MRP < 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "MRP cannot be negative",
			})
			return
		}

		if m.MinimumStock < 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Minimum Stock cannot be negative",
			})
			return
		}

		m.ID = id

		query := `
			UPDATE medicines
			SET
				name=$1,
				strength=$2,
				generic=$3,
				category_id=$4,
				type_id=$5,
				box_size_id=$6,
				unit_id=$7,
				leaf_id=$8,
				selling_price=$9,
				mrp=$10,
				status=$11,
				updated_at=CURRENT_TIMESTAMP
			WHERE id=$12
			RETURNING
				current_stock,
				created_at,
				updated_at
		`

		err = db.QueryRow(
			query,
			m.Name,
			m.Strength,
			m.Generic,
			m.CategoryID,
			m.TypeID,
			m.BoxSizeID,
			m.UnitID,
			m.LeafID,
			m.SellingPrice,
			m.MRP,
			m.Status,
			m.ID,
		).Scan(
			&m.CurrentStock,
			&m.CreatedAt,
			&m.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, Response{
				Status:  false,
				Message: "Not Found",
				Error:   "Medicine not found",
			})
			return
		}

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Database Error",
				Error:   "Failed to update medicine: " + err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, Response{
			Status:  true,
			Message: "Medicine updated successfully",
			Data:    m,
		})
	}
}

// -------------------------------------------------------------
// 5. DELETE MEDICINE (SOFT DELETE)
// DELETE /api/v1/medicines/{id}
// -------------------------------------------------------------
func DeleteMedicineHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			writeJSON(w, http.StatusMethodNotAllowed, Response{
				Status:  false,
				Message: "Method not allowed",
				Error:   "Only DELETE method is allowed for this endpoint",
			})
			return
		}

		id, err := parseIDFromURL(r)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Status:  false,
				Message: "Validation Error",
				Error:   "Invalid medicine ID",
			})
			return
		}

		query := `
			UPDATE medicines
			SET
				status = 0,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND status = 1
		`

		result, err := db.Exec(query, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Database Error",
				Error:   "Failed to delete medicine: " + err.Error(),
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, Response{
				Status:  false,
				Message: "Database Error",
				Error:   err.Error(),
			})
			return
		}

		if rowsAffected == 0 {
			writeJSON(w, http.StatusNotFound, Response{
				Status:  false,
				Message: "Not Found",
				Error:   "Medicine not found or already deleted",
			})
			return
		}

		writeJSON(w, http.StatusOK, Response{
			Status:  true,
			Message: "Medicine deleted successfully",
		})
	}
}
