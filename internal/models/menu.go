package models

type MenuItem struct {
	ID       int        `json:"id"`
	ParentID *int       `json:"parent_id,omitempty"`
	Path     *string    `json:"path,omitempty"`
	Label    string     `json:"label"`
	Icon     *string    `json:"icon,omitempty"`
	Roles    []string   `json:"roles"`
	Children []MenuItem `json:"children,omitempty"`
}
