package router

import (
	"database/sql"
	"meditrack-backend/internal/handlers/purchase"
	"net/http"
)

func RegisterPurchaseRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	// 1. Create Purchase
	RegisterRoute(mux, prefix, "POST /purchases", purchase.CreatePurchase(db))

	// 2. Get All Purchases List
	RegisterRoute(mux, prefix, "GET /purchases", purchase.GetPurchases(db))

	// 3. Get Single Purchase Details by ID
	RegisterRoute(mux, prefix, "GET /purchases/{id}", purchase.GetPurchaseByID(db))

	// 4. Update Purchase Invoice
	RegisterRoute(mux, prefix, "PUT /purchases/{id}", purchase.UpdatePurchase(db))

	// 5. Soft Delete Purchase Invoice (with Delete Reason)
	RegisterRoute(mux, prefix, "DELETE /purchases/{id}", purchase.DeletePurchase(db))

	// 6. Process Purchase Return (Invoice-wise & Item-wise)
	RegisterRoute(mux, prefix, "POST /purchases/return", purchase.ProcessPurchaseReturn(db))
}
