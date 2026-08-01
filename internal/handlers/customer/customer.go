package customer

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
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateCustomerRequest struct {
	CustomerName    string  `json:"customerName"`
	Mobile          string  `json:"mobile"`
	Email           string  `json:"email"`
	ContactPerson   string  `json:"contactPerson"`
	Address         string  `json:"address"`
	City            string  `json:"city"`
	State           string  `json:"state"`
	Zip             string  `json:"zip"`
	Country         string  `json:"country"`
	PreviousBalance float64 `json:"previousBalance"`
}

type UpdateCustomerRequest struct {
	CustomerName   string  `json:"customerName"`
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

func respondJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

func parseIDFromURL(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = path.Base(r.URL.Path)
	}
	return strconv.Atoi(idStr)
}

func validateCustomer(name, mobile string) error {
	name = strings.TrimSpace(name)
	mobile = strings.TrimSpace(mobile)

	if name == "" {
		return errors.New("Customer name is required")
	}

	if mobile == "" {
		return errors.New("Mobile number is required")
	}

	return nil
}

func CreateCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req CreateCustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid request payload",
			})
			return
		}

		if strings.TrimSpace(req.CustomerName) == "" {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Customer name is required",
			})
			return
		}

		if strings.TrimSpace(req.Mobile) == "" {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Mobile number is required",
			})
			return
		}

		query := `
			INSERT INTO customers (
				customer_name, mobile, email, contact_person, address, 
				city, state, zip, country, opening_balance, current_balance, created_at, updated_at
			) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id, created_at
		`

		var createdID int
		var createdAt time.Time
		now := time.Now()

		country := req.Country
		if country == "" {
			country = "Bangladesh"
		}

		err := db.QueryRowContext(
			ctx, query,
			req.CustomerName, req.Mobile, req.Email, req.ContactPerson, req.Address,
			req.City, req.State, req.Zip, country, req.PreviousBalance, req.PreviousBalance, now, now,
		).Scan(&createdID, &createdAt)

		if err != nil {
			log.Println("Error inserting customer:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to create customer",
			})
			return
		}

		respondJSON(w, http.StatusCreated, Response{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Customer added successfully",
			Data: map[string]interface{}{
				"id":           createdID,
				"customerName": req.CustomerName,
				"mobile":       req.Mobile,
				"createdAt":    createdAt.Format(time.RFC3339),
			},
		})
	}
}

func GetCustomers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := `
			SELECT id, customer_name, mobile, email, contact_person, address, 
			       city, state, zip, country, opening_balance, current_balance, status, created_at
			FROM customers 
			ORDER BY id DESC
		`

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Println("Error fetching customers:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to fetch customers",
			})
			return
		}
		defer rows.Close()

		customerList := make([]map[string]interface{}, 0)

		for rows.Next() {
			var id, status int
			var name, mobile, email, contactPerson, address, city, state, zip, country string
			var openingBal, currentBal float64
			var createdAt time.Time

			err := rows.Scan(
				&id, &name, &mobile, &email, &contactPerson, &address,
				&city, &state, &zip, &country, &openingBal, &currentBal, &status, &createdAt,
			)
			if err != nil {
				log.Println("Scan error:", err)
				continue
			}

			customerList = append(customerList, map[string]interface{}{
				"id":             id,
				"customerName":   name,
				"mobile":         mobile,
				"email":          email,
				"contactPerson":  contactPerson,
				"address":        address,
				"city":           city,
				"state":          state,
				"zip":            zip,
				"country":        country,
				"openingBalance": openingBal,
				"currentBalance": currentBal,
				"status":         status,
				"createdAt":      createdAt.Format(time.RFC3339),
			})
		}

		if err := rows.Err(); err != nil {
			log.Println("Error during row iteration:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Error processing customer rows",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "customers retrieved successfully",
			Data:    customerList,
		})
	}
}

func GetCustomerByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil || id <= 0 {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid customer ID",
			})
			return
		}

		query := `
			SELECT id, customer_name, mobile, email, contact_person, address, 
			       city, state, zip, country, opening_balance, current_balance, status, created_at, updated_at
			FROM customers 
			WHERE id = $1
		`

		var s map[string]interface{}
		var customerID, status int
		var name, mobile, email, contactPerson, address, city, state, zip, country string
		var openingBal, currentBal float64
		var createdAt, updatedAt time.Time

		err = db.QueryRowContext(ctx, query, id).Scan(
			&customerID, &name, &mobile, &email, &contactPerson, &address,
			&city, &state, &zip, &country, &openingBal, &currentBal, &status, &createdAt, &updatedAt,
		)

		if err == sql.ErrNoRows {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "customer not found",
			})
			return
		} else if err != nil {
			log.Println("Error fetching customer by ID:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to fetch customer details",
			})
			return
		}

		s = map[string]interface{}{
			"id":             customerID,
			"customerName":   name,
			"mobile":         mobile,
			"email":          email,
			"contactPerson":  contactPerson,
			"address":        address,
			"city":           city,
			"state":          state,
			"zip":            zip,
			"country":        country,
			"openingBalance": openingBal,
			"currentBalance": currentBal,
			"status":         status,
			"createdAt":      createdAt.Format(time.RFC3339),
			"updatedAt":      updatedAt.Format(time.RFC3339),
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "customer details fetched successfully",
			Data:    s,
		})
	}
}

func UpdateCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil || id <= 0 {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid customer ID",
			})
			return
		}

		var req UpdateCustomerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid JSON request body",
			})
			return
		}

		// SQL UPDATE Query
		query := `
			UPDATE customers 
			SET customer_name = $1,
			    mobile = $2,
			    email = $3,
			    contact_person = $4,
			    address = $5,
			    city = $6,
			    state = $7,
			    zip = $8,
			    country = $9,
			    opening_balance = $10,
			    status = $11,
			    updated_at = $12
			WHERE id = $13
		`

		now := time.Now()

		result, err := db.ExecContext(ctx, query,
			req.CustomerName,
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
			now,
			id,
		)

		if err != nil {
			log.Println("Error updating customer:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to update customer",
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Customer not found or no changes made",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Customer updated successfully",
		})
	}
}

func DeleteCustomer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil || id <= 0 {
			respondJSON(w, http.StatusBadRequest, Response{
				Status:  http.StatusBadRequest,
				Success: false,
				Message: "Invalid customer ID",
			})
			return
		}

		query := `UPDATE customers SET status = 0, updated_at = $1 WHERE id = $2`

		now := time.Now()
		result, err := db.ExecContext(ctx, query, now, id)
		if err != nil {
			log.Println("Error deactivating customer:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Status:  http.StatusInternalServerError,
				Success: false,
				Message: "Failed to deactivate customer",
			})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, Response{
				Status:  http.StatusNotFound,
				Success: false,
				Message: "Customer not found",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Status:  http.StatusOK,
			Success: true,
			Message: "Customer deactivated successfully",
		})
	}
}
