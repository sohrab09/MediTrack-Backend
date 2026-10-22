package router

import (
	"database/sql"
	"meditrack-backend/global_router"
	"net/http"
)

func SetupRoutes(mux *http.ServeMux, db *sql.DB) http.Handler {
	apiPrefix := "/api/v1"

	// Health Check
	RegisterRoute(mux, apiPrefix, "GET /health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Feature-wise Route Group Integration
	RegisterMenuRoutes(mux, apiPrefix, db)
	RegisterAuthRoutes(mux, apiPrefix, db)
	RegisterUserRoutes(mux, apiPrefix, db)
	RegisterMedicineRoutes(mux, apiPrefix, db)
	RegisterSupplierRoutes(mux, apiPrefix, db)
	RegisterCustomerRoutes(mux, apiPrefix, db)
	RegisterPurchaseRoutes(mux, apiPrefix, db)
	RegisterBankRoutes(mux, apiPrefix, db)

	// Global Middleware Setup
	return global_router.GlobalRouter(mux)
}
