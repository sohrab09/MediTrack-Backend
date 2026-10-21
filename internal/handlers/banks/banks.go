package banks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"meditrack-backend/internal/models"
)

// Standard API Response Structure
type Response struct {
	Status  bool        `json:"status"`          // true for success, false for error
	Code    int         `json:"code"`            // HTTP Status Code (200, 400, 500, etc.)
	Message string      `json:"message"`         // Human-readable message
	Data    interface{} `json:"data,omitempty"`  // Data payload (omitted if nil)
	Error   string      `json:"error,omitempty"` // Error details (omitted if empty)
}

// Helper function for JSON Response
func respondJSON(w http.ResponseWriter, httpCode int, success bool, message string, data interface{}, errDetails string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(Response{
		Status:  success,
		Code:    httpCode,
		Message: message,
		Data:    data,
		Error:   errDetails,
	})
}

// 1. Create Bank Account
func CreateBank(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var b models.Bank

		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			respondJSON(w, http.StatusBadRequest, false, "Invalid request payload", nil, err.Error())
			return
		}

		if strings.TrimSpace(b.BankName) == "" || strings.TrimSpace(b.AccountNumber) == "" {
			respondJSON(w, http.StatusBadRequest, false, "Bank name and account number are required", nil, "Validation Error: Missing required fields")
			return
		}

		b.CurrentBalance = b.OpeningBalance
		if b.Status == 0 {
			b.Status = 1
		}

		query := `
			INSERT INTO banks (
				bank_name, account_holder_name, account_number, account_type, 
				branch_name, routing_number, opening_balance, current_balance, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, created_at, updated_at
		`

		err := db.QueryRowContext(
			ctx, query,
			b.BankName, b.AccountHolderName, b.AccountNumber, b.AccountType,
			b.BranchName, b.RoutingNumber, b.OpeningBalance, b.CurrentBalance, b.Status,
		).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)

		if err != nil {
			respondJSON(w, http.StatusInternalServerError, false, "Failed to create bank account", nil, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, true, "Bank account created successfully", b, "")
	}
}

// 2. Get All Bank Accounts
func GetBanks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rows, err := db.QueryContext(ctx, "SELECT id, bank_name, account_holder_name, account_number, account_type, branch_name, routing_number, opening_balance, current_balance, status, created_at, updated_at FROM banks ORDER BY id DESC")
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, false, "Failed to fetch banks", nil, err.Error())
			return
		}
		defer rows.Close()

		var banks []models.Bank
		for rows.Next() {
			var b models.Bank
			if err := rows.Scan(&b.ID, &b.BankName, &b.AccountHolderName, &b.AccountNumber, &b.AccountType, &b.BranchName, &b.RoutingNumber, &b.OpeningBalance, &b.CurrentBalance, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
				continue
			}
			banks = append(banks, b)
		}

		if err := rows.Err(); err != nil {
			respondJSON(w, http.StatusInternalServerError, false, "Error reading bank rows", nil, err.Error())
			return
		}

		if banks == nil {
			banks = []models.Bank{}
		}

		respondJSON(w, http.StatusOK, true, "Banks retrieved successfully", banks, "")
	}
}

// 3. GetBankByID - Retrieve single bank account details
func GetBankByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			respondJSON(w, http.StatusBadRequest, false, "Invalid bank ID", nil, "Validation Error: ID must be a positive integer")
			return
		}

		query := `SELECT id, bank_name, account_holder_name, account_number, account_type, branch_name, routing_number, opening_balance, current_balance, status, created_at, updated_at FROM banks WHERE id = $1`

		var b models.Bank
		err = db.QueryRowContext(ctx, query, id).Scan(
			&b.ID, &b.BankName, &b.AccountHolderName, &b.AccountNumber, &b.AccountType,
			&b.BranchName, &b.RoutingNumber, &b.OpeningBalance, &b.CurrentBalance,
			&b.Status, &b.CreatedAt, &b.UpdatedAt,
		)

		if errors.Is(err, sql.ErrNoRows) {
			respondJSON(w, http.StatusNotFound, false, "Bank account not found", nil, err.Error())
			return
		} else if err != nil {
			respondJSON(w, http.StatusInternalServerError, false, "Database error", nil, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, true, "Bank retrieved successfully", b, "")
	}
}

// 4. Update Bank Account
func UpdateBank(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			respondJSON(w, http.StatusBadRequest, false, "Invalid bank ID", nil, "Validation Error: ID must be a positive integer")
			return
		}

		var b models.Bank
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			respondJSON(w, http.StatusBadRequest, false, "Invalid request payload", nil, err.Error())
			return
		}

		query := `
			UPDATE banks SET 
				bank_name = $1, 
				account_holder_name = $2, 
				account_number = $3, 
				account_type = $4, 
				branch_name = $5, 
				routing_number = $6, 
				status = $7,
				updated_at = NOW()
			WHERE id = $8
			RETURNING id, bank_name, account_holder_name, account_number, account_type, branch_name, routing_number, opening_balance, current_balance, status, created_at, updated_at
		`

		err = db.QueryRowContext(
			ctx, query,
			b.BankName, b.AccountHolderName, b.AccountNumber, b.AccountType,
			b.BranchName, b.RoutingNumber, b.Status, id,
		).Scan(&b.ID, &b.BankName, &b.AccountHolderName, &b.AccountNumber, &b.AccountType, &b.BranchName, &b.RoutingNumber, &b.OpeningBalance, &b.CurrentBalance, &b.Status, &b.CreatedAt, &b.UpdatedAt)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondJSON(w, http.StatusNotFound, false, "Bank account not found", nil, err.Error())
				return
			}
			respondJSON(w, http.StatusInternalServerError, false, "Failed to update bank account", nil, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, true, "Bank account updated successfully", b, "")
	}
}

// 5. Delete Bank Account
func DeleteBank(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			respondJSON(w, http.StatusBadRequest, false, "Invalid bank ID", nil, "Validation Error: ID must be a positive integer")
			return
		}

		result, err := db.ExecContext(ctx, "DELETE FROM banks WHERE id = $1", id)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, false, "Failed to delete bank account", nil, err.Error())
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, false, "Bank account not found", nil, "No record found with the given ID")
			return
		}

		respondJSON(w, http.StatusOK, true, "Bank account deleted successfully", nil, "")
	}
}
