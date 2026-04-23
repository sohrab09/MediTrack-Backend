package deletemedicinecategories

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

func DeleteMedicineCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodDelete {
			respondJSON(w, http.StatusMethodNotAllowed, Response{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)

		if err != nil || id <= 0 {
			respondJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid ID",
			})
			return
		}

		result, err := db.Exec("DELETE FROM categories WHERE id = $1", id)
		if err != nil {
			log.Println("Delete error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to delete category",
			})
			return
		}

		rowsAffected, _ := result.RowsAffected()

		if rowsAffected == 0 {
			respondJSON(w, http.StatusNotFound, Response{
				Success: false,
				Message: "Category not found",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Success: true,
			Message: "Category deleted successfully",
		})
	}
}
