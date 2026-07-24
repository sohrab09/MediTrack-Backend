package router

import (
	"database/sql"
	"net/http"

	"meditrack-backend/internal/handlers/login"
	"meditrack-backend/internal/handlers/register"
)

func RegisterAuthRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	RegisterRoute(mux, prefix, "POST /auth/login", login.LoginHandler(db))
	RegisterRoute(mux, prefix, "POST /auth/register", register.RegisterHandler(db))
}
