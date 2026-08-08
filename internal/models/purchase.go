package models

import "time"

type Purchase struct {
	ID        int    `json:"id"`
	InvoiceNo string `json:"invoice_no"`

	SupplierID   int    `json:"supplier_id"`
	SupplierName string `json:"supplier_name,omitempty"`

	PurchaseDate time.Time `json:"purchase_date"`

	SubTotal float64 `json:"subtotal"`
	Discount float64 `json:"discount"`
	Tax      float64 `json:"tax"`

	TotalAmount float64 `json:"total_amount"`

	PaidAmount float64 `json:"paid_amount"`
	DueAmount  float64 `json:"due_amount"`

	PaymentStatus string `json:"payment_status"`

	Note *string `json:"note,omitempty"`

	Status int `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Items []PurchaseItem `json:"items,omitempty"`
}

type PurchaseItem struct {
	ID int `json:"id"`

	PurchaseID int `json:"purchase_id"`

	MedicineID   int    `json:"medicine_id"`
	MedicineName string `json:"medicine_name,omitempty"`

	BatchNo string `json:"batch_no"`

	ExpiryDate time.Time `json:"expiry_date"`

	Quantity int `json:"quantity"`

	PurchasePrice float64 `json:"purchase_price"`

	SalePrice float64 `json:"sale_price"`

	TotalPrice float64 `json:"total_price"`

	Status int `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
