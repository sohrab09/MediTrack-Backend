package router

import (
	"database/sql"
	"meditrack-backend/internal/handlers/customer"
	"net/http"
)

func RegisterCustomerRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	RegisterRoute(mux, prefix, "POST /customers", customer.CreateCustomer(db))
	RegisterRoute(mux, prefix, "GET /customers", customer.GetCustomers(db))
	RegisterRoute(mux, prefix, "GET /customers/{id}", customer.GetCustomerByID(db))
	RegisterRoute(mux, prefix, "PUT /customers/{id}", customer.UpdateCustomer(db))
	RegisterRoute(mux, prefix, "DELETE /customers/{id}", customer.DeleteCustomer(db))
}
