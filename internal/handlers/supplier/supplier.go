package supplier

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Status  bool        `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateSupplierRequest struct {
	SupplierName   string  `json:"supplierName"`
	Mobile         string  `json:"mobile"`
	Email          string  `json:"email"`
	ContactPerson  string  `json:"contactPerson"`
	Address        string  `json:"address"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Zip            string  `json:"zip"`
	Country        string  `json:"country"`
	OpeningBalance float64 `json:"openingBalance"`
}

type UpdateSupplierRequest struct {
	SupplierName   string  `json:"supplierName"`
	Mobile         string  `json:"mobile"`
	Email          string  `json:"email"`
	ContactPerson  string  `json:"contactPerson"`
	Address        string  `json:"address"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Zip            string  `json:"zip"`
	Country        string  `json:"country"`
	OpeningBalance float64 `json:"openingBalance"`
	Status         int     `json:"status"`
}

func respondJSON(w http.ResponseWriter, statusCode int, res Response) {
	res.Code = statusCode

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("Failed to encode response:", err)
	}
}

func parseIDFromURL(r *http.Request) (int, error) {
	idStr := r.PathValue("id")

	if idStr == "" {
		idStr = path.Base(r.URL.Path)
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid supplier id")
	}

	return id, nil
}

func validateSupplier(name, mobile string) error {
	name = strings.TrimSpace(name)
	mobile = strings.TrimSpace(mobile)

	if name == "" {
		return errors.New("supplier name is required")
	}

	if mobile == "" {
		return errors.New("mobile number is required")
	}

	return nil
}

func respondError(w http.ResponseWriter, statusCode int, message string, err error) {
	res := Response{
		Status:  false,
		Message: message,
	}

	if err != nil {
		res.Error = err.Error()
	}

	respondJSON(w, statusCode, res)
}

func respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	respondJSON(w, statusCode, Response{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

func CreateSupplier(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req CreateSupplierRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				"Invalid request payload",
				err,
			)
			return
		}

		// Trim Input
		req.SupplierName = strings.TrimSpace(req.SupplierName)
		req.Mobile = strings.TrimSpace(req.Mobile)
		req.Email = strings.TrimSpace(req.Email)
		req.ContactPerson = strings.TrimSpace(req.ContactPerson)
		req.Address = strings.TrimSpace(req.Address)
		req.City = strings.TrimSpace(req.City)
		req.State = strings.TrimSpace(req.State)
		req.Zip = strings.TrimSpace(req.Zip)
		req.Country = strings.TrimSpace(req.Country)

		// Validation
		if err := validateSupplier(req.SupplierName, req.Mobile); err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				err.Error(),
				err,
			)
			return
		}

		if req.Country == "" {
			req.Country = "Bangladesh"
		}

		query := `
			INSERT INTO suppliers (
				supplier_name,
				mobile,
				email,
				contact_person,
				address,
				city,
				state,
				zip,
				country,
				opening_balance,
				current_balance,
				created_at,
				updated_at
			)
			VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),NOW()
			)
			RETURNING id, created_at
		`

		var supplierID int
		var createdAt time.Time

		err := db.QueryRowContext(
			ctx,
			query,
			req.SupplierName,
			req.Mobile,
			req.Email,
			req.ContactPerson,
			req.Address,
			req.City,
			req.State,
			req.Zip,
			req.Country,
			req.OpeningBalance,
			req.OpeningBalance,
		).Scan(
			&supplierID,
			&createdAt,
		)

		if err != nil {
			log.Println("CreateSupplier:", err)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to create supplier",
				err,
			)
			return
		}

		respondSuccess(
			w,
			http.StatusCreated,
			"Supplier created successfully",
			map[string]interface{}{
				"id":             supplierID,
				"supplierName":   req.SupplierName,
				"mobile":         req.Mobile,
				"openingBalance": req.OpeningBalance,
				"currentBalance": req.OpeningBalance,
				"createdAt":      createdAt.Format(time.RFC3339),
			},
		)
	}
}

func GetSuppliers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := `
			SELECT
				id,
				supplier_name,
				mobile,
				email,
				contact_person,
				address,
				city,
				state,
				zip,
				country,
				opening_balance,
				current_balance,
				status,
				created_at,
				updated_at
			FROM suppliers
			ORDER BY id DESC
		`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Println("GetSuppliers:", err)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to fetch suppliers",
				err,
			)
			return
		}
		defer rows.Close()

		suppliers := make([]map[string]interface{}, 0)

		for rows.Next() {

			var (
				id           int
				supplierName string
				mobile       string

				email         sql.NullString
				contactPerson sql.NullString
				address       sql.NullString
				city          sql.NullString
				state         sql.NullString
				zip           sql.NullString
				country       sql.NullString

				openingBalance float64
				currentBalance float64
				status         int

				createdAt time.Time
				updatedAt time.Time
			)

			err := rows.Scan(
				&id,
				&supplierName,
				&mobile,
				&email,
				&contactPerson,
				&address,
				&city,
				&state,
				&zip,
				&country,
				&openingBalance,
				&currentBalance,
				&status,
				&createdAt,
				&updatedAt,
			)

			if err != nil {
				log.Println("GetSuppliers Scan:", err)
				continue
			}

			suppliers = append(suppliers, map[string]interface{}{
				"id":             id,
				"supplierName":   supplierName,
				"mobile":         mobile,
				"email":          email.String,
				"contactPerson":  contactPerson.String,
				"address":        address.String,
				"city":           city.String,
				"state":          state.String,
				"zip":            zip.String,
				"country":        country.String,
				"openingBalance": openingBalance,
				"currentBalance": currentBalance,
				"status":         status,
				"createdAt":      createdAt.Format(time.RFC3339),
				"updatedAt":      updatedAt.Format(time.RFC3339),
			})
		}

		if err := rows.Err(); err != nil {
			log.Println("GetSuppliers Rows:", err)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to process suppliers",
				err,
			)
			return
		}

		respondSuccess(
			w,
			http.StatusOK,
			"Suppliers retrieved successfully",
			suppliers,
		)
	}
}

func GetSupplierByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				"Invalid supplier ID",
				err,
			)
			return
		}

		query := `
			SELECT
				id,
				supplier_name,
				mobile,
				email,
				contact_person,
				address,
				city,
				state,
				zip,
				country,
				opening_balance,
				current_balance,
				status,
				created_at,
				updated_at
			FROM suppliers
			WHERE id = $1
		`

		var (
			supplierID   int
			supplierName string
			mobile       string

			email         sql.NullString
			contactPerson sql.NullString
			address       sql.NullString
			city          sql.NullString
			state         sql.NullString
			zip           sql.NullString
			country       sql.NullString

			openingBalance float64
			currentBalance float64
			status         int

			createdAt time.Time
			updatedAt time.Time
		)

		err = db.QueryRowContext(ctx, query, id).Scan(
			&supplierID,
			&supplierName,
			&mobile,
			&email,
			&contactPerson,
			&address,
			&city,
			&state,
			&zip,
			&country,
			&openingBalance,
			&currentBalance,
			&status,
			&createdAt,
			&updatedAt,
		)

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				respondError(
					w,
					http.StatusNotFound,
					"Supplier not found",
					err,
				)
				return
			}

			log.Println("GetSupplierByID:", err)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to fetch supplier",
				err,
			)
			return
		}

		supplier := map[string]interface{}{
			"id":             supplierID,
			"supplierName":   supplierName,
			"mobile":         mobile,
			"email":          email.String,
			"contactPerson":  contactPerson.String,
			"address":        address.String,
			"city":           city.String,
			"state":          state.String,
			"zip":            zip.String,
			"country":        country.String,
			"openingBalance": openingBalance,
			"currentBalance": currentBalance,
			"status":         status,
			"createdAt":      createdAt.Format(time.RFC3339),
			"updatedAt":      updatedAt.Format(time.RFC3339),
		}

		respondSuccess(
			w,
			http.StatusOK,
			"Supplier retrieved successfully",
			supplier,
		)
	}
}

func UpdateSupplier(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				"Invalid supplier ID",
				err,
			)
			return
		}

		var req UpdateSupplierRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				"Invalid request payload",
				err,
			)
			return
		}

		// Trim Input
		req.SupplierName = strings.TrimSpace(req.SupplierName)
		req.Mobile = strings.TrimSpace(req.Mobile)
		req.Email = strings.TrimSpace(req.Email)
		req.ContactPerson = strings.TrimSpace(req.ContactPerson)
		req.Address = strings.TrimSpace(req.Address)
		req.City = strings.TrimSpace(req.City)
		req.State = strings.TrimSpace(req.State)
		req.Zip = strings.TrimSpace(req.Zip)
		req.Country = strings.TrimSpace(req.Country)

		// Validation
		if err := validateSupplier(req.SupplierName, req.Mobile); err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				err.Error(),
				err,
			)
			return
		}

		if req.Country == "" {
			req.Country = "Bangladesh"
		}

		query := `
			UPDATE suppliers
			SET
				supplier_name    = $1,
				mobile           = $2,
				email            = $3,
				contact_person   = $4,
				address          = $5,
				city             = $6,
				state            = $7,
				zip              = $8,
				country          = $9,
				opening_balance  = $10,
				status           = $11,
				updated_at       = NOW()
			WHERE id = $12
		`

		result, err := db.ExecContext(
			ctx,
			query,
			req.SupplierName,
			req.Mobile,
			req.Email,
			req.ContactPerson,
			req.Address,
			req.City,
			req.State,
			req.Zip,
			req.Country,
			req.OpeningBalance,
			req.Status,
			id,
		)

		if err != nil {
			log.Println("UpdateSupplier:", err)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to update supplier",
				err,
			)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("UpdateSupplier RowsAffected:", err)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to update supplier",
				err,
			)
			return
		}

		if rowsAffected == 0 {
			respondError(
				w,
				http.StatusNotFound,
				"Supplier not found",
				errors.New("supplier not found"),
			)
			return
		}

		respondSuccess(
			w,
			http.StatusOK,
			"Supplier updated successfully",
			nil,
		)
	}
}

func DeleteSupplier(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				"Invalid supplier ID",
				err,
			)
			return
		}

		query := `
			UPDATE suppliers
			SET
				status = 0,
				updated_at = NOW()
			WHERE id = $1
		`

		result, err := db.ExecContext(ctx, query, id)
		if err != nil {
			log.Println("DeleteSupplier:", err)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to delete supplier",
				err,
			)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("DeleteSupplier RowsAffected:", err)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to delete supplier",
				err,
			)
			return
		}

		if rowsAffected == 0 {
			respondError(
				w,
				http.StatusNotFound,
				"Supplier not found",
				errors.New("supplier not found"),
			)
			return
		}

		respondSuccess(
			w,
			http.StatusOK,
			"Supplier deleted successfully",
			nil,
		)
	}
}
