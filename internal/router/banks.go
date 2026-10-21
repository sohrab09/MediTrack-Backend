package router

import (
	"database/sql"
	"meditrack-backend/internal/handlers/banks"
	"net/http"
)

func RegisterBankRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	// 1. Create Bank Account
	RegisterRoute(mux, prefix, "POST /banks", banks.CreateBank(db))

	// 2. Get All Bank Accounts
	RegisterRoute(mux, prefix, "GET /banks", banks.GetBanks(db))

	// 3. Get Single Bank Account by ID
	RegisterRoute(mux, prefix, "GET /banks/{id}", banks.GetBankByID(db))

	// 4. Update Bank Account Information
	RegisterRoute(mux, prefix, "PUT /banks/{id}", banks.UpdateBank(db))

	// 5. Delete Bank Account
	RegisterRoute(mux, prefix, "DELETE /banks/{id}", banks.DeleteBank(db))
}
