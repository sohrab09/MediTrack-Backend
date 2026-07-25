package router

import (
	"database/sql"
	"meditrack-backend/internal/handlers/menu"
	"net/http"
)

func RegisterMenuRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	RegisterRoute(mux, prefix, "GET /menus", menu.GetMenusHandler(db))
}
