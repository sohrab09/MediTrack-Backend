package updatemedicinecategories

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

type UpdateRequest struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

func UpdateMedicineCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPut {
			respondJSON(w, http.StatusMethodNotAllowed, Response{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		// URL: /categories?id=1
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			respondJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid ID",
			})
			return
		}

		var req UpdateRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Message: "Invalid JSON",
			})
			return
		}

		if strings.TrimSpace(req.Name) == "" {
			respondJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Message: "Name is required",
			})
			return
		}

		query := `
			UPDATE categories
			SET name = $1, status = $2
			WHERE id = $3
			RETURNING id, name, status, created_at
		`

		type Category struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Status    int    `json:"status"`
			CreatedAt string `json:"created_at"`
		}

		var cat Category

		err = db.QueryRow(query, req.Name, req.Status, id).
			Scan(&cat.ID, &cat.Name, &cat.Status, &cat.CreatedAt)

		if err != nil {
			log.Println("Update error:", err)
			respondJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to update category",
			})
			return
		}

		respondJSON(w, http.StatusOK, Response{
			Success: true,
			Message: "Category updated successfully",
			Data:    cat,
		})
	}
}
