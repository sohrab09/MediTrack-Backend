package router

import (
	"database/sql"
	"net/http"

	"meditrack-backend/global_router"
)

func SetupRoutes(mux *http.ServeMux, db *sql.DB, jwtSecret string) http.Handler {
	apiPrefix := "/api/v1"

	// Health Check
	RegisterRoute(mux, apiPrefix, "GET /health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Feature-wise Route Group Integration
	RegisterMenuRoutes(mux, apiPrefix, db)
	RegisterAuthRoutes(mux, apiPrefix, db, jwtSecret)
	RegisterUserRoutes(mux, apiPrefix, db)
	RegisterMedicineRoutes(mux, apiPrefix, db)
	RegisterSupplierRoutes(mux, apiPrefix, db)
	RegisterCustomerRoutes(mux, apiPrefix, db)
	RegisterPurchaseRoutes(mux, apiPrefix, db)
	RegisterBankRoutes(mux, apiPrefix, db)

	// Global Middleware Setup
	return global_router.GlobalRouter(mux)
}
