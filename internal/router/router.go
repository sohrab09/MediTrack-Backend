package router

import (
	"database/sql"
	"net/http"

	"meditrack-backend/router"
)

func SetupRoutes(mux *http.ServeMux, db *sql.DB) http.Handler {
	apiPrefix := "/api/v1"

	// Health Check
	RegisterRoute(mux, apiPrefix, "GET /health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Feature-wise Route Group Integration
	RegisterAuthRoutes(mux, apiPrefix, db)
	RegisterUserRoutes(mux, apiPrefix, db)
	RegisterMedicineRoutes(mux, apiPrefix, db)

	// Global Middleware Setup
	return router.GlobalRouter(mux)
}
