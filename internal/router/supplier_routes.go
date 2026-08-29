package router

import (
	"database/sql"
	suppliers "meditrack-backend/internal/handlers/supplier"
	"net/http"
)

func RegisterSupplierRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	RegisterRoute(mux, prefix, "POST /suppliers", suppliers.CreateSupplier(db))
	RegisterRoute(mux, prefix, "GET /suppliers", suppliers.GetSuppliers(db))
	RegisterRoute(mux, prefix, "GET /suppliers/{id}", suppliers.GetSupplierByID(db))
	RegisterRoute(mux, prefix, "PUT /suppliers/{id}", suppliers.UpdateSupplier(db))
	RegisterRoute(mux, prefix, "DELETE /suppliers/{id}", suppliers.DeleteSupplier(db))
}
