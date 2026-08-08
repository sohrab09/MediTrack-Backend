package router

import (
	"database/sql"
	"meditrack-backend/internal/handlers/purchase"
	"net/http"
)

func RegisterPurchaseRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	RegisterRoute(mux, prefix, "POST /purchases", purchase.CreatePurchase(db))
	RegisterRoute(mux, prefix, "GET /purchases", purchase.GetPurchases(db))
	RegisterRoute(mux, prefix, "GET /purchases/{id}", purchase.GetPurchaseByID(db))
	RegisterRoute(mux, prefix, "PUT /purchases/{id}", purchase.UpdatePurchase(db))
	RegisterRoute(mux, prefix, "DELETE /purchases/{id}", purchase.DeletePurchase(db))
}
