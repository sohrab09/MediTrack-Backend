package router

import (
	"database/sql"
	"meditrack-backend/internal/handlers/supplier"
	"net/http"
)

func RegisterSupplierRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	RegisterRoute(mux, prefix, "POST /suppliers", supplier.CreateSupplier(db))
	RegisterRoute(mux, prefix, "GET /suppliers", supplier.GetSuppliers(db))
	RegisterRoute(mux, prefix, "GET /suppliers/{id}", supplier.GetSupplierByID(db))
	RegisterRoute(mux, prefix, "PUT /suppliers/{id}", supplier.UpdateSupplier(db))
	RegisterRoute(mux, prefix, "DELETE /suppliers/{id}", supplier.DeleteSupplier(db))
}
