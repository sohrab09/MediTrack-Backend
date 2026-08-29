package suppliers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strconv"
	"strings"

	"meditrack-backend/internal/models"
)

// Standard Response Structure
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

// URL/Query থেকে Supplier ID এক্সট্র্যাক্ট করা
func parseIDFromURL(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	if idStr == "" {
		idStr = path.Base(r.URL.Path)
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid supplier id")
	}

	return id, nil
}

// Supplier Input Validation
func validateSupplier(name, mobile string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("supplier name is required")
	}
	if strings.TrimSpace(mobile) == "" {
		return errors.New("mobile number is required")
	}
	return nil
}

// CreateSupplier - Add a new supplier record
func CreateSupplier(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var s models.Supplier

		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Input Validation
		if err := validateSupplier(s.SupplierName, s.Mobile); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		if strings.TrimSpace(s.Country) == "" {
			s.Country = "Bangladesh"
		}

		// JWT Context থেকে User ID বের করা
		s.CreatedBy = getUserIDFromContext(ctx)
		s.CurrentBalance = s.OpeningBalance
		if s.Status == 0 {
			s.Status = 1
		}

		query := `
			INSERT INTO suppliers (
				supplier_name, mobile, email, contact_person, address, 
				city, state, zip, country, opening_balance, current_balance, 
				status, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id, created_at, updated_at
		`

		err := db.QueryRowContext(
			ctx, query,
			s.SupplierName, s.Mobile, s.Email, s.ContactPerson, s.Address,
			s.City, s.State, s.Zip, s.Country, s.OpeningBalance, s.CurrentBalance,
			s.Status, s.CreatedBy,
		).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create supplier: "+err.Error())
			return
		}

		respondSuccess(w, http.StatusCreated, "Supplier created successfully", s)
	}
}

// GetSuppliers - Retrieve supplier list
func GetSuppliers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := `
			SELECT id, supplier_name, mobile, email, contact_person, address, 
			       city, state, zip, country, opening_balance, current_balance, 
			       status, created_by, created_at, updated_at
			FROM suppliers ORDER BY id DESC
		`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Database error: "+err.Error())
			return
		}
		defer rows.Close()

		suppliers := make([]models.Supplier, 0)

		for rows.Next() {
			var s models.Supplier
			err := rows.Scan(
				&s.ID, &s.SupplierName, &s.Mobile, &s.Email, &s.ContactPerson, &s.Address,
				&s.City, &s.State, &s.Zip, &s.Country, &s.OpeningBalance, &s.CurrentBalance,
				&s.Status, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
			)

			if err != nil {
				respondError(w, http.StatusInternalServerError, "Error scanning row: "+err.Error())
				return
			}

			suppliers = append(suppliers, s)
		}

		if err := rows.Err(); err != nil {
			respondError(w, http.StatusInternalServerError, "Error iterating suppliers: "+err.Error())
			return
		}

		respondSuccess(w, http.StatusOK, "Suppliers retrieved successfully", suppliers)
	}
}

// GetSupplierByID - Retrieve single supplier details
func GetSupplierByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid supplier ID")
			return
		}

		query := `
			SELECT id, supplier_name, mobile, email, contact_person, address, 
			       city, state, zip, country, opening_balance, current_balance, 
			       status, created_by, created_at, updated_at
			FROM suppliers WHERE id = $1
		`

		var s models.Supplier
		err = db.QueryRowContext(ctx, query, id).Scan(
			&s.ID, &s.SupplierName, &s.Mobile, &s.Email, &s.ContactPerson, &s.Address,
			&s.City, &s.State, &s.Zip, &s.Country, &s.OpeningBalance, &s.CurrentBalance,
			&s.Status, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		)

		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Supplier not found")
			return
		} else if err != nil {
			respondError(w, http.StatusInternalServerError, "Database error: "+err.Error())
			return
		}

		respondSuccess(w, http.StatusOK, "Supplier retrieved successfully", s)
	}
}

// UpdateSupplier - Update supplier info
func UpdateSupplier(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid supplier ID")
			return
		}

		var s models.Supplier
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if err := validateSupplier(s.SupplierName, s.Mobile); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		// SQL Query তে opening_balance এবং current_balance যুক্ত করা হয়েছে
		query := `
			UPDATE suppliers
			SET supplier_name = $1, mobile = $2, email = $3, contact_person = $4, 
				address = $5, city = $6, state = $7, zip = $8, country = $9, 
				opening_balance = $10, current_balance = $11, status = $12, 
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $13
			RETURNING opening_balance, current_balance, updated_at
		`

		err = db.QueryRowContext(
			ctx, query,
			s.SupplierName, s.Mobile, s.Email, s.ContactPerson,
			s.Address, s.City, s.State, s.Zip, s.Country,
			s.OpeningBalance, s.OpeningBalance, // opening_balance আপডেট করার সাথে সাথে current_balance ও আপডেট করা হচ্ছে
			s.Status, id,
		).Scan(&s.OpeningBalance, &s.CurrentBalance, &s.UpdatedAt)

		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Supplier not found")
			return
		} else if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update supplier: "+err.Error())
			return
		}

		s.ID = id
		respondSuccess(w, http.StatusOK, "Supplier updated successfully", s)
	}
}

// DeleteSupplier - Soft Delete supplier (Status = 0)
func DeleteSupplier(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid supplier ID")
			return
		}

		query := `UPDATE suppliers SET status = 0, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

		result, err := db.ExecContext(ctx, query, id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to delete supplier: "+err.Error())
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			respondError(w, http.StatusNotFound, "Supplier not found")
			return
		}

		respondSuccess(w, http.StatusOK, "Supplier deleted successfully", nil)
	}
}
