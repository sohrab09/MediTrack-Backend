package router

import (
	"database/sql"
	"net/http"

	"meditrack-backend/internal/handlers/medicine_category"
	"meditrack-backend/internal/handlers/medicine_leaves"
	"meditrack-backend/internal/handlers/medicine_types"
	"meditrack-backend/internal/handlers/medicine_units"
)

func RegisterMedicineRoutes(mux *http.ServeMux, prefix string, db *sql.DB) {
	// Medicine Categories
	RegisterRoute(mux, prefix, "POST /medicine-categories", medicine_category.CreateMedicineCategories(db))
	RegisterRoute(mux, prefix, "GET /medicine-categories", medicine_category.GetMedicineCategories(db))
	RegisterRoute(mux, prefix, "GET /medicine-categories/{id}", medicine_category.GetMedicineCategoryByID(db))
	RegisterRoute(mux, prefix, "PUT /medicine-categories/{id}", medicine_category.UpdateMedicineCategory(db))
	RegisterRoute(mux, prefix, "DELETE /medicine-categories/{id}", medicine_category.DeleteMedicineCategory(db))

	// Medicine Units
	RegisterRoute(mux, prefix, "POST /medicine-units", medicine_units.CreateMedicineUnits(db))
	RegisterRoute(mux, prefix, "GET /medicine-units", medicine_units.GetMedicineUnits(db))
	RegisterRoute(mux, prefix, "GET /medicine-units/{id}", medicine_units.GetMedicineUnitByID(db))
	RegisterRoute(mux, prefix, "PUT /medicine-units/{id}", medicine_units.UpdateMedicineUnits(db))
	RegisterRoute(mux, prefix, "DELETE /medicine-units/{id}", medicine_units.DeleteMedicineUnits(db))

	// Medicine Types
	RegisterRoute(mux, prefix, "POST /medicine-types", medicine_types.CreateMedicineTypes(db))
	RegisterRoute(mux, prefix, "GET /medicine-types", medicine_types.GetMedicineTypes(db))
	RegisterRoute(mux, prefix, "GET /medicine-types/{id}", medicine_types.GetMedicineTypesByID(db))
	RegisterRoute(mux, prefix, "PUT /medicine-types/{id}", medicine_types.UpdateMedicineTypes(db))
	RegisterRoute(mux, prefix, "DELETE /medicine-types/{id}", medicine_types.DeleteMedicineTypes(db))

	// Medicine Leafs
	RegisterRoute(mux, prefix, "POST /medicine-leaves", medicine_leaves.CreateMedicineLeaf(db))
	RegisterRoute(mux, prefix, "GET /medicine-leaves", medicine_leaves.GetMedicineLeaves(db))
	RegisterRoute(mux, prefix, "GET /medicine-leaves/{id}", medicine_leaves.GetMedicineLeafByID(db))
	RegisterRoute(mux, prefix, "PUT /medicine-leaves/{id}", medicine_leaves.UpdateMedicineLeaf(db))
	RegisterRoute(mux, prefix, "DELETE /medicine-leaves/{id}", medicine_leaves.DeleteMedicineLeaf(db))
}
