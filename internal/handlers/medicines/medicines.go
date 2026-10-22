package medicines

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"meditrack-backend/internal/models"
)

type Response struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type medicineInput struct {
	// ID is optional for backward compatibility with existing frontend requests.
	// The URL path ID remains the source of truth during updates.
	ID           int     `json:"id,omitempty"`
	Name         string  `json:"name"`
	Generic      *string `json:"generic"`
	Strength     *string `json:"strength"`
	UnitID       int     `json:"unit_id"`
	CategoryID   int     `json:"category_id"`
	LeafID       *int    `json:"leaf_id"`
	TypeID       *int    `json:"type_id"`
	BoxSizeID    int     `json:"box_size_id"`
	MinimumStock int     `json:"minimum_stock"`
	Status       *int    `json:"status"`
}

func respondSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(Response{
		Status:  status,
		Success: true,
		Message: message,
		Data:    data,
	})
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(Response{
		Status:  status,
		Success: false,
		Message: message,
	})
}

func decodeMedicineInput(r *http.Request) (medicineInput, error) {
	var input medicineInput

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		return input, err
	}

	// Reject multiple JSON objects or any unexpected trailing content.
	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return input, errors.New("request must contain one JSON object")
		}
		return input, err
	}

	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		return input, errors.New("name is required")
	}

	if input.UnitID <= 0 || input.CategoryID <= 0 || input.BoxSizeID <= 0 {
		return input, errors.New("unit_id, category_id and box_size_id must be positive")
	}

	if input.LeafID != nil && *input.LeafID <= 0 {
		return input, errors.New("leaf_id must be positive when provided")
	}

	if input.TypeID != nil && *input.TypeID <= 0 {
		return input, errors.New("type_id must be positive when provided")
	}

	if input.MinimumStock < 0 {
		return input, errors.New("minimum_stock cannot be negative")
	}

	if input.Status != nil && *input.Status != 0 && *input.Status != 1 {
		return input, errors.New("status must be 0 or 1")
	}

	return input, nil
}

// medicineID reads an ID from the URL path or the ?id= query parameter.
func medicineID(r *http.Request, fromPath bool) (int, error) {
	idStr := r.URL.Query().Get("id")

	if fromPath {
		idStr = r.PathValue("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid medicine ID")
	}

	return id, nil
}

func applyMedicineInput(m *models.Medicine, input medicineInput) {
	m.Name = input.Name
	m.Generic = input.Generic
	m.Strength = input.Strength
	m.UnitID = input.UnitID
	m.CategoryID = input.CategoryID
	m.LeafID = input.LeafID
	m.TypeID = input.TypeID
	m.BoxSizeID = input.BoxSizeID
	m.MinimumStock = input.MinimumStock
}

// AddMedicine creates a medicine with status = 1 and current_stock = 0.
func AddMedicine(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := decodeMedicineInput(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid medicine data: "+err.Error())
			return
		}

		query := `
			INSERT INTO medicines (
				name,
				generic,
				strength,
				unit_id,
				category_id,
				leaf_id,
				type_id,
				box_size_id,
				minimum_stock,
				current_stock,
				status
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0, 1)
			RETURNING id, created_at, updated_at
		`

		var m models.Medicine
		applyMedicineInput(&m, input)

		err = db.QueryRowContext(
			r.Context(),
			query,
			input.Name,
			input.Generic,
			input.Strength,
			input.UnitID,
			input.CategoryID,
			input.LeafID,
			input.TypeID,
			input.BoxSizeID,
			input.MinimumStock,
		).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)

		if err != nil {
			log.Printf("AddMedicine: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to add medicine")
			return
		}

		m.Status = 1
		respondSuccess(w, http.StatusCreated, "Medicine added successfully", m)
	}
}

// GetMedicines
func GetMedicines(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
			SELECT
				id,
				name,
				generic,
				strength,
				unit_id,
				category_id,
				leaf_id,
				type_id,
				box_size_id,
				minimum_stock,
				current_stock,
				status,
				created_at,
				updated_at
			FROM medicines
			ORDER BY id DESC;
		`

		rows, err := db.QueryContext(r.Context(), query)
		if err != nil {
			log.Printf("GetMedicines: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to fetch medicines")
			return
		}
		defer rows.Close()

		list := make([]models.Medicine, 0)

		for rows.Next() {
			var m models.Medicine

			if err := rows.Scan(
				&m.ID,
				&m.Name,
				&m.Generic,
				&m.Strength,
				&m.UnitID,
				&m.CategoryID,
				&m.LeafID,
				&m.TypeID,
				&m.BoxSizeID,
				&m.MinimumStock,
				&m.CurrentStock,
				&m.Status,
				&m.CreatedAt,
				&m.UpdatedAt,
			); err != nil {
				log.Printf("GetMedicines scan: %v", err)
				respondError(w, http.StatusInternalServerError, "Failed to read medicines")
				return
			}

			list = append(list, m)
		}

		if err := rows.Err(); err != nil {
			log.Printf("GetMedicines rows: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to read medicines")
			return
		}

		respondSuccess(w, http.StatusOK, "Medicines fetched successfully", list)
	}
}

// GetMedicineByID
func GetMedicineByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := medicineID(r, false)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid or missing medicine ID")
			return
		}

		query := `
			SELECT
				id,
				name,
				generic,
				strength,
				unit_id,
				category_id,
				leaf_id,
				type_id,
				box_size_id,
				minimum_stock,
				current_stock,
				status,
				created_at,
				updated_at
			FROM medicines
			WHERE id = $1
		`

		var m models.Medicine

		err = db.QueryRowContext(r.Context(), query, id).Scan(
			&m.ID,
			&m.Name,
			&m.Generic,
			&m.Strength,
			&m.UnitID,
			&m.CategoryID,
			&m.LeafID,
			&m.TypeID,
			&m.BoxSizeID,
			&m.MinimumStock,
			&m.CurrentStock,
			&m.Status,
			&m.CreatedAt,
			&m.UpdatedAt,
		)

		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Medicine not found")
			return
		}

		if err != nil {
			log.Printf("GetMedicineByID: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to fetch medicine")
			return
		}

		respondSuccess(w, http.StatusOK, "Medicine details fetched", m)
	}
}

// UpdateMedicine
func UpdateMedicine(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := medicineID(r, true)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid or missing medicine ID")
			return
		}

		input, err := decodeMedicineInput(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid medicine data: "+err.Error())
			return
		}

		// If the client also sends an ID in the body, it must match the URL.
		if input.ID != 0 && input.ID != id {
			respondError(w, http.StatusBadRequest, "Medicine ID in request body does not match URL ID")
			return
		}

		var statusArg interface{}
		if input.Status != nil {
			statusArg = *input.Status
		}

		query := `
			UPDATE medicines
			SET
				name = $1,
				generic = $2,
				strength = $3,
				unit_id = $4,
				category_id = $5,
				leaf_id = $6,
				type_id = $7,
				box_size_id = $8,
				minimum_stock = $9,
				status = COALESCE($10, status),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $11
			RETURNING id, current_stock, status, created_at, updated_at
		`

		var m models.Medicine
		applyMedicineInput(&m, input)

		err = db.QueryRowContext(
			r.Context(),
			query,
			input.Name,
			input.Generic,
			input.Strength,
			input.UnitID,
			input.CategoryID,
			input.LeafID,
			input.TypeID,
			input.BoxSizeID,
			input.MinimumStock,
			statusArg,
			id,
		).Scan(
			&m.ID,
			&m.CurrentStock,
			&m.Status,
			&m.CreatedAt,
			&m.UpdatedAt,
		)

		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Medicine not found")
			return
		}

		if err != nil {
			log.Printf("UpdateMedicine: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to update medicine")
			return
		}

		respondSuccess(w, http.StatusOK, "Medicine updated successfully", m)
	}
}

// DeleteMedicine
func DeleteMedicine(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := medicineID(r, true)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid or missing medicine ID")
			return
		}

		query := `
			UPDATE medicines
			SET
				status = 0,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`

		result, err := db.ExecContext(r.Context(), query, id)
		if err != nil {
			log.Printf("DeleteMedicine: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to change medicine status")
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("DeleteMedicine rows affected: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to change medicine status")
			return
		}

		if rowsAffected == 0 {
			respondError(w, http.StatusNotFound, "Medicine not found")
			return
		}

		respondSuccess(w, http.StatusOK, "Medicine status changed to inactive", nil)
	}
}
