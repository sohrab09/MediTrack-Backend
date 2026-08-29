package models

import "time"

type PurchaseItem struct {
	ID            int       `json:"id"`
	PurchaseID    int       `json:"purchase_id"`
	MedicineID    int       `json:"medicine_id"`
	BatchNo       string    `json:"batch_no"`
	ExpiryDate    *string   `json:"expiry_date,omitempty"`
	Quantity      int       `json:"quantity"`
	PurchasePrice float64   `json:"purchase_price"`
	SellingPrice  float64   `json:"selling_price"`
	TotalPrice    float64   `json:"total_price"`
	CreatedAt     time.Time `json:"created_at"`
}

type Purchase struct {
	ID            int            `json:"id"`
	InvoiceNo     string         `json:"invoice_no"`
	SupplierID    int            `json:"supplier_id"`
	UserID        int            `json:"user_id"`
	PurchaseDate  time.Time      `json:"purchase_date"`
	Subtotal      float64        `json:"subtotal"`
	Discount      float64        `json:"discount"`
	Tax           float64        `json:"tax"`
	TotalAmount   float64        `json:"total_amount"`
	PaidAmount    float64        `json:"paid_amount"`
	DueAmount     float64        `json:"due_amount"`
	PaymentStatus string         `json:"payment_status"`
	Note          *string        `json:"note,omitempty"`
	Status        int            `json:"status"` // 1 = Active, 0 = Deleted
	DeletedBy     *int           `json:"deleted_by,omitempty"`
	DeletedAt     *time.Time     `json:"deleted_at,omitempty"`
	DeleteReason  *string        `json:"delete_reason,omitempty"`
	UpdatedBy     *int           `json:"updated_by,omitempty"`
	Items         []PurchaseItem `json:"items,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Purchase Return Models
type PurchaseReturnItem struct {
	ID               int       `json:"id"`
	PurchaseReturnID int       `json:"purchase_return_id"`
	PurchaseItemID   int       `json:"purchase_item_id"`
	MedicineID       int       `json:"medicine_id"`
	ReturnQty        int       `json:"return_qty"`
	UnitPrice        float64   `json:"unit_price"`
	TotalPrice       float64   `json:"total_price"`
	ItemReason       *string   `json:"item_reason,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type PurchaseReturn struct {
	ID                int                  `json:"id"`
	PurchaseID        int                  `json:"purchase_id"`
	SupplierID        int                  `json:"supplier_id"`
	UserID            int                  `json:"user_id"`
	ReturnType        string               `json:"return_type"` // "invoice_wise" or "item_wise"
	TotalReturnAmount float64              `json:"total_return_amount"`
	ReturnReason      string               `json:"return_reason"`
	Items             []PurchaseReturnItem `json:"items,omitempty"`
	CreatedAt         time.Time            `json:"created_at"`
}
