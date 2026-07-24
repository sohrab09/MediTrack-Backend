package router

import (
	"database/sql"
	"net/http"

	"meditrack-backend/internal/handlers/users"
)

func RegisterUserRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	RegisterRoute(mux, prefix, "GET /users", users.GetUsers(db))
	RegisterRoute(mux, prefix, "GET /users/{id}", users.GetUser(db))
	RegisterRoute(mux, prefix, "PUT /users/{id}", users.UpdateUser(db))
	RegisterRoute(mux, prefix, "DELETE /users/{id}", users.DeleteUser(db))
}
