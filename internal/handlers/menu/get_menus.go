package menu

import (
	"database/sql"
	"encoding/json"
	"meditrack-backend/internal/models"
	"net/http"

	"github.com/lib/pq"
)

func GetMenusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := `SELECT id, parent_id, path, label, icon, roles FROM menus ORDER BY id ASC`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, `{"error": "Failed to fetch menus"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		menuMap := make(map[int]*models.MenuItem)
		var menuIDs []int

		for rows.Next() {
			var item models.MenuItem
			var roles pq.StringArray
			var parentID *int
			var path, icon *string

			err := rows.Scan(&item.ID, &parentID, &path, &item.Label, &icon, &roles)
			if err != nil {
				http.Error(w, `{"error": "Error scanning menu data"}`, http.StatusInternalServerError)
				return
			}

			item.ParentID = parentID
			item.Path = path
			item.Icon = icon
			item.Roles = roles
			item.Children = []models.MenuItem{}

			itemCopy := item
			menuMap[item.ID] = &itemCopy
			menuIDs = append(menuIDs, item.ID)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, `{"error": "Error iterating menu rows"}`, http.StatusInternalServerError)
			return
		}

		var rootMenuPointers []*models.MenuItem

		for _, id := range menuIDs {
			item := menuMap[id]

			if item.ParentID == nil {
				rootMenuPointers = append(rootMenuPointers, item)
			} else {
				if parent, exists := menuMap[*item.ParentID]; exists {
					parent.Children = append(parent.Children, *item)
				}
			}
		}

		var rootMenus []models.MenuItem
		for _, ptr := range rootMenuPointers {
			rootMenus = append(rootMenus, *ptr)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    rootMenus,
		})
	}
}
