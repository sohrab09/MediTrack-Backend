package purchase

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Status  bool        `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PurchaseItemRequest struct {
	MedicineID    int     `json:"medicineID"`
	BatchNo       string  `json:"batchNo"`
	ExpiryDate    string  `json:"expiryDate"`
	Quantity      int     `json:"quantity"`
	PurchasePrice float64 `json:"purchasePrice"`
	SalePrice     float64 `json:"salePrice"`
}

type CreatePurchaseRequest struct {
	InvoiceNo string `json:"invoiceNo"`

	SupplierID int `json:"supplierID"`

	PurchaseDate string `json:"purchaseDate"`

	SubTotal float64 `json:"subTotal"`

	Discount float64 `json:"discount"`

	Tax float64 `json:"tax"`

	TotalAmount float64 `json:"totalAmount"`

	PaidAmount float64 `json:"paidAmount"`

	DueAmount float64 `json:"dueAmount"`

	PaymentStatus string `json:"paymentStatus"`

	Note string `json:"note"`

	Items []PurchaseItemRequest `json:"items"`
}

type UpdatePurchaseRequest struct {
	ID int `json:"id"`

	CreatePurchaseRequest
}

func respondJSON(w http.ResponseWriter, statusCode int, res Response) {
	res.Code = statusCode

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("Failed to encode response:", err)
	}
}

func respondError(
	w http.ResponseWriter,
	statusCode int,
	message string,
	err error,
) {
	res := Response{
		Status:  false,
		Message: message,
	}

	if err != nil {
		res.Error = err.Error()
	}

	respondJSON(w, statusCode, res)
}

func respondSuccess(
	w http.ResponseWriter,
	statusCode int,
	message string,
	data interface{},
) {
	respondJSON(w, statusCode, Response{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

func parseIDFromURL(r *http.Request) (int, error) {

	idStr := r.PathValue("id")

	if idStr == "" {
		idStr = path.Base(r.URL.Path)
	}

	id, err := strconv.Atoi(idStr)

	if err != nil || id <= 0 {
		return 0, errors.New("invalid purchase id")
	}

	return id, nil
}

func validatePurchase(req CreatePurchaseRequest) error {

	if req.SupplierID <= 0 {
		return errors.New("supplier is required")
	}

	if strings.TrimSpace(req.InvoiceNo) == "" {
		return errors.New("invoice number is required")
	}

	if len(req.Items) == 0 {
		return errors.New("at least one medicine is required")
	}

	for _, item := range req.Items {

		if item.MedicineID <= 0 {
			return errors.New("invalid medicine")
		}

		if item.Quantity <= 0 {
			return errors.New("quantity must be greater than zero")
		}

		if item.PurchasePrice <= 0 {
			return errors.New("purchase price must be greater than zero")
		}

		if item.SalePrice < 0 {
			return errors.New("sale price must be greater than zero")
		}
	}

	return nil
}

func CreatePurchase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		var req CreatePurchaseRequest

		// ============================
		// DECODE REQUEST BODY
		// ============================

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				"Invalid request payload",
				err,
			)
			return
		}

		// ============================
		// TRIM INPUT
		// ============================

		req.InvoiceNo = strings.TrimSpace(req.InvoiceNo)
		req.Note = strings.TrimSpace(req.Note)
		req.PaymentStatus = strings.TrimSpace(req.PaymentStatus)

		// ============================
		// DEFAULT PAYMENT STATUS
		// ============================

		switch {
		case req.PaidAmount <= 0:
			req.PaymentStatus = "unpaid"

		case req.DueAmount <= 0:
			req.PaymentStatus = "paid"

		default:
			req.PaymentStatus = "partial"
		}

		// ============================
		// VALIDATE REQUEST
		// ============================

		if err := validatePurchase(req); err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				err.Error(),
				err,
			)
			return
		}

		// ============================
		// START TRANSACTION
		// ============================

		tx, err := db.BeginTx(ctx, nil)

		if err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to start transaction",
				err,
			)
			return
		}

		defer tx.Rollback()

		var purchaseID int

		// ============================
		// INSERT PURCHASE
		// ============================

		purchaseQuery := `
			INSERT INTO purchases
			(
				invoice_no,
				supplier_id,
				purchase_date,

				subtotal,
				discount,
				tax,

				total_amount,

				paid_amount,
				due_amount,

				payment_status,

				note,

				status,

				created_at,
				updated_at
			)
			VALUES
			(
				$1,
				$2,
				$3,

				$4,
				$5,
				$6,

				$7,

				$8,
				$9,

				$10,

				$11,

				1,

				NOW(),
				NOW()
			)
			RETURNING id
		`

		err = tx.QueryRowContext(
			ctx,
			purchaseQuery,

			req.InvoiceNo,
			req.SupplierID,
			req.PurchaseDate,

			req.SubTotal,
			req.Discount,
			req.Tax,

			req.TotalAmount,

			req.PaidAmount,
			req.DueAmount,

			req.PaymentStatus,

			req.Note,
		).Scan(&purchaseID)

		if err != nil {
			log.Println(
				"Create Purchase:",
				err,
			)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to create purchase",
				err,
			)
			return
		}

		// ============================
		// INSERT PURCHASE ITEMS
		// ============================

		itemQuery := `
			INSERT INTO purchase_items
			(
				purchase_id,
				medicine_id,
				batch_no,
				expiry_date,
				quantity,
				purchase_price,
				selling_price,
				total_price,
				created_at
			)
			VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				NOW()
			)
		`

		// ============================
		// UPDATE MEDICINE STOCK
		// ============================

		updateMedicineQuery := `
			UPDATE medicines
			SET
				current_stock = current_stock + $1,
				selling_price = $2,
				mrp = $3,
				updated_at = NOW()
			WHERE id = $4
		`

		for _, item := range req.Items {

			// Calculate item total
			totalPrice := float64(item.Quantity) * item.PurchasePrice

			// ============================
			// INSERT PURCHASE ITEM
			// ============================

			_, err = tx.ExecContext(
				ctx,
				itemQuery,

				purchaseID,
				item.MedicineID,
				item.BatchNo,
				item.ExpiryDate,
				item.Quantity,
				item.PurchasePrice,
				item.SalePrice,
				totalPrice,
			)

			if err != nil {
				log.Println(
					"Insert Purchase Item:",
					err,
				)

				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to save purchase items",
					err,
				)
				return
			}

			// ============================
			// UPDATE MEDICINE STOCK
			// ============================

			_, err = tx.ExecContext(
				ctx,
				updateMedicineQuery,

				item.Quantity,

				// selling_price
				item.SalePrice,

				// mrp
				item.SalePrice,

				// medicine id
				item.MedicineID,
			)

			if err != nil {
				log.Println(
					"Update Medicine Stock:",
					err,
				)

				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to update medicine stock",
					err,
				)
				return
			}
		}

		// ============================
		// UPDATE SUPPLIER BALANCE
		// ============================

		updateSupplierQuery := `
			UPDATE suppliers
			SET
				current_balance = current_balance + $1,
				updated_at = NOW()
			WHERE id = $2
		`

		_, err = tx.ExecContext(
			ctx,
			updateSupplierQuery,

			req.DueAmount,
			req.SupplierID,
		)

		if err != nil {
			log.Println(
				"Update Supplier Balance:",
				err,
			)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to update supplier balance",
				err,
			)
			return
		}

		// ============================
		// COMMIT TRANSACTION
		// ============================

		if err := tx.Commit(); err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to complete purchase",
				err,
			)
			return
		}

		// ============================
		// SUCCESS RESPONSE
		// ============================

		respondSuccess(
			w,
			http.StatusCreated,
			"Purchase created successfully",
			map[string]interface{}{
				"purchaseID":    purchaseID,
				"invoiceNo":     req.InvoiceNo,
				"supplierID":    req.SupplierID,
				"totalAmount":   req.TotalAmount,
				"dueAmount":     req.DueAmount,
				"paymentStatus": req.PaymentStatus,
			},
		)
	}
}

func GetPurchases(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		query := `

			SELECT

				p.id,

				p.invoice_no,

				p.supplier_id,

				s.supplier_name,


				p.purchase_date,


				p.subtotal,

				p.discount,

				p.tax,


				p.total_amount,


				p.paid_amount,

				p.due_amount,


				p.payment_status,


				p.status,


				p.created_at,

				p.updated_at


			FROM purchases p


			LEFT JOIN suppliers s

			ON p.supplier_id = s.id


			ORDER BY p.id DESC

		`

		rows, err := db.QueryContext(ctx, query)

		if err != nil {

			log.Println(
				"GetPurchases:",
				err,
			)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to fetch purchases",
				err,
			)

			return
		}

		defer rows.Close()

		purchases := make([]map[string]interface{}, 0)

		for rows.Next() {

			var (
				id int

				supplierID int

				invoiceNo string

				supplierName sql.NullString

				purchaseDate time.Time

				subtotal float64

				discount float64

				tax float64

				totalAmount float64

				paidAmount float64

				dueAmount float64

				paymentStatus string

				status int

				createdAt time.Time

				updatedAt time.Time
			)

			err := rows.Scan(

				&id,

				&invoiceNo,

				&supplierID,

				&supplierName,

				&purchaseDate,

				&subtotal,

				&discount,

				&tax,

				&totalAmount,

				&paidAmount,

				&dueAmount,

				&paymentStatus,

				&status,

				&createdAt,

				&updatedAt,
			)

			if err != nil {

				log.Println(
					"GetPurchases Scan:",
					err,
				)

				continue
			}

			purchases = append(

				purchases,

				map[string]interface{}{

					"id": id,

					"invoiceNo": invoiceNo,

					"supplierID": supplierID,

					"supplierName": supplierName.String,

					"purchaseDate": purchaseDate.Format(time.RFC3339),

					"subtotal": subtotal,

					"discount": discount,

					"tax": tax,

					"totalAmount": totalAmount,

					"paidAmount": paidAmount,

					"dueAmount": dueAmount,

					"paymentStatus": paymentStatus,

					"status": status,

					"createdAt": createdAt.Format(time.RFC3339),

					"updatedAt": updatedAt.Format(time.RFC3339),
				},
			)

		}

		if err := rows.Err(); err != nil {

			log.Println(
				"GetPurchases Rows:",
				err,
			)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to process purchases",
				err,
			)

			return
		}

		respondSuccess(

			w,

			http.StatusOK,

			"Purchases retrieved successfully",

			purchases,
		)

	}

}

func GetPurchaseByID(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		// ==========================
		// Parse Purchase ID
		// ==========================

		id, err := parseIDFromURL(r)

		if err != nil {

			respondError(
				w,
				http.StatusBadRequest,
				"Invalid purchase ID",
				err,
			)

			return
		}

		// ==========================
		// Purchase Header
		// ==========================

		purchaseQuery := `
			SELECT
				p.id,
				p.invoice_no,

				p.supplier_id,

				s.supplier_name,
				s.mobile,
				s.address,

				p.purchase_date,

				p.subtotal,
				p.discount,
				p.tax,

				p.total_amount,

				p.paid_amount,
				p.due_amount,

				p.payment_status,

				p.note,

				p.status,

				p.created_at,
				p.updated_at

			FROM purchases p

			LEFT JOIN suppliers s
				ON p.supplier_id = s.id

			WHERE p.id = $1
		`

		var (
			purchaseID int

			invoiceNo string

			supplierID int

			supplierName    sql.NullString
			supplierMobile  sql.NullString
			supplierAddress sql.NullString

			purchaseDate time.Time

			subTotal float64
			discount float64
			tax      float64

			totalAmount float64

			paidAmount float64
			dueAmount  float64

			paymentStatus string

			note sql.NullString

			status int

			createdAt time.Time
			updatedAt time.Time
		)

		err = db.QueryRowContext(
			ctx,
			purchaseQuery,
			id,
		).Scan(

			&purchaseID,

			&invoiceNo,

			&supplierID,

			&supplierName,
			&supplierMobile,
			&supplierAddress,

			&purchaseDate,

			&subTotal,
			&discount,
			&tax,

			&totalAmount,

			&paidAmount,
			&dueAmount,

			&paymentStatus,

			&note,

			&status,

			&createdAt,
			&updatedAt,
		)

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {

				respondError(
					w,
					http.StatusNotFound,
					"Purchase not found",
					err,
				)

				return
			}

			log.Println(
				"GetPurchaseByID Header:",
				err,
			)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to fetch purchase",
				err,
			)

			return
		}

		// ==========================
		// Purchase Response
		// ==========================

		purchase := map[string]interface{}{

			"id": purchaseID,

			"invoiceNo": invoiceNo,

			"supplier": map[string]interface{}{

				"id": supplierID,

				"name": supplierName.String,

				"mobile": supplierMobile.String,

				"address": supplierAddress.String,
			},

			"purchaseDate": purchaseDate.Format(
				time.RFC3339,
			),

			"subtotal": subTotal,

			"discount": discount,

			"tax": tax,

			"totalAmount": totalAmount,

			"paidAmount": paidAmount,

			"dueAmount": dueAmount,

			"paymentStatus": paymentStatus,

			"note": note.String,

			"status": status,

			"createdAt": createdAt.Format(
				time.RFC3339,
			),

			"updatedAt": updatedAt.Format(
				time.RFC3339,
			),
		}

		// ==========================
		// Purchase Items
		// ==========================

		itemQuery := `
			SELECT
				pi.id,
				pi.medicine_id,

				m.code,
				m.barcode,
				m.name,
				m.strength,
				m.generic,

				pi.batch_no,
				pi.expiry_date,

				pi.quantity,

				pi.purchase_price,
				pi.selling_price,
				pi.total_price

			FROM purchase_items pi

			INNER JOIN medicines m
				ON pi.medicine_id = m.id

			WHERE pi.purchase_id = $1

			ORDER BY pi.id ASC
		`

		rows, err := db.QueryContext(
			ctx,
			itemQuery,
			purchaseID,
		)

		if err != nil {

			log.Println(
				"GetPurchaseByID Items:",
				err,
			)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to fetch purchase items",
				err,
			)

			return
		}

		defer rows.Close()

		// ==========================
		// Purchase Items List
		// ==========================

		items := make(
			[]map[string]interface{},
			0,
		)

		for rows.Next() {

			var (
				itemID int

				medicineID int

				medicineCode     sql.NullString
				medicineBarcode  sql.NullString
				medicineName     sql.NullString
				medicineStrength sql.NullString
				medicineGeneric  sql.NullString

				batchNo    sql.NullString
				expiryDate sql.NullTime

				quantity int

				purchasePrice float64
				sellingPrice  float64
				totalPrice    float64
			)

			err := rows.Scan(

				&itemID,

				&medicineID,

				&medicineCode,
				&medicineBarcode,
				&medicineName,
				&medicineStrength,
				&medicineGeneric,

				&batchNo,
				&expiryDate,

				&quantity,

				&purchasePrice,
				&sellingPrice,
				&totalPrice,
			)

			if err != nil {

				log.Println(
					"GetPurchaseByID Item Scan:",
					err,
				)

				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to read purchase item",
					err,
				)

				return
			}

			// ==========================
			// Format Nullable Expiry Date
			// ==========================

			var formattedExpiryDate interface{}

			if expiryDate.Valid {

				formattedExpiryDate =
					expiryDate.Time.Format(
						time.RFC3339,
					)

			} else {

				formattedExpiryDate = nil
			}

			// ==========================
			// Append Item
			// ==========================

			items = append(
				items,
				map[string]interface{}{

					"id": itemID,

					"medicineID": medicineID,

					"medicine": map[string]interface{}{

						"code": medicineCode.String,

						"barcode": medicineBarcode.String,

						"name": medicineName.String,

						"strength": medicineStrength.String,

						"generic": medicineGeneric.String,
					},

					"batchNo": batchNo.String,

					"expiryDate": formattedExpiryDate,

					"quantity": quantity,

					"purchasePrice": purchasePrice,

					"sellingPrice": sellingPrice,

					"totalPrice": totalPrice,
				},
			)
		}

		// ==========================
		// Check Rows Error
		// ==========================

		if err := rows.Err(); err != nil {

			log.Println(
				"GetPurchaseByID Rows:",
				err,
			)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to read purchase items",
				err,
			)

			return
		}

		// ==========================
		// Attach Items
		// ==========================

		purchase["items"] = items

		// ==========================
		// Success Response
		// ==========================

		respondSuccess(

			w,

			http.StatusOK,

			"Purchase details retrieved successfully",

			purchase,
		)
	}
}

func UpdatePurchase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		// ======================================
		// Get Purchase ID from URL
		// ======================================

		id, err := parseIDFromURL(r)

		if err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				"Invalid purchase ID",
				err,
			)
			return
		}

		// ======================================
		// Decode Request
		// ======================================

		var req CreatePurchaseRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				"Invalid request payload",
				err,
			)
			return
		}

		fmt.Printf("========== UPDATE PURCHASE ==========\n")
		fmt.Printf("Purchase ID: %d\n", id)
		fmt.Printf("Supplier ID: %d\n", req.SupplierID)
		fmt.Printf("Invoice No: %s\n", req.InvoiceNo)
		fmt.Printf("Items: %+v\n", req.Items)
		fmt.Printf("=====================================\n")

		// ======================================
		// Trim String Fields
		// ======================================

		req.InvoiceNo = strings.TrimSpace(req.InvoiceNo)
		req.Note = strings.TrimSpace(req.Note)

		// ======================================
		// Basic Validation
		// ======================================

		if err := validatePurchase(req); err != nil {
			respondError(
				w,
				http.StatusBadRequest,
				err.Error(),
				err,
			)
			return
		}

		// ======================================
		// Validate Items
		// ======================================

		if len(req.Items) == 0 {
			respondError(
				w,
				http.StatusBadRequest,
				"Purchase must contain at least one item",
				nil,
			)
			return
		}

		for i, item := range req.Items {

			if item.MedicineID <= 0 {
				respondError(
					w,
					http.StatusBadRequest,
					fmt.Sprintf("Invalid medicine ID at item %d", i+1),
					nil,
				)
				return
			}

			if item.Quantity <= 0 {
				respondError(
					w,
					http.StatusBadRequest,
					fmt.Sprintf("Quantity must be greater than zero at item %d", i+1),
					nil,
				)
				return
			}

			if item.PurchasePrice < 0 {
				respondError(
					w,
					http.StatusBadRequest,
					fmt.Sprintf("Purchase price cannot be negative at item %d", i+1),
					nil,
				)
				return
			}

			if item.SalePrice < 0 {
				respondError(
					w,
					http.StatusBadRequest,
					fmt.Sprintf("Sale price cannot be negative at item %d", i+1),
					nil,
				)
				return
			}

			item.BatchNo = strings.TrimSpace(item.BatchNo)

			if item.BatchNo == "" {
				respondError(
					w,
					http.StatusBadRequest,
					fmt.Sprintf("Batch number is required at item %d", i+1),
					nil,
				)
				return
			}

			if strings.TrimSpace(item.ExpiryDate) == "" {
				respondError(
					w,
					http.StatusBadRequest,
					fmt.Sprintf("Expiry date is required at item %d", i+1),
					nil,
				)
				return
			}
		}

		// ======================================
		// Calculate Financial Values
		// ======================================

		var calculatedSubTotal float64

		for _, item := range req.Items {

			totalPrice := float64(item.Quantity) * item.PurchasePrice

			calculatedSubTotal += totalPrice
		}

		// Discount validation
		if req.Discount < 0 {
			respondError(
				w,
				http.StatusBadRequest,
				"Discount cannot be negative",
				nil,
			)
			return
		}

		if req.Discount > calculatedSubTotal {
			respondError(
				w,
				http.StatusBadRequest,
				"Discount cannot be greater than subtotal",
				nil,
			)
			return
		}

		// Tax validation
		if req.Tax < 0 {
			respondError(
				w,
				http.StatusBadRequest,
				"Tax cannot be negative",
				nil,
			)
			return
		}

		// Calculate total
		calculatedTotal := calculatedSubTotal - req.Discount + req.Tax

		// Paid amount validation
		if req.PaidAmount < 0 {
			respondError(
				w,
				http.StatusBadRequest,
				"Paid amount cannot be negative",
				nil,
			)
			return
		}

		if req.PaidAmount > calculatedTotal {
			respondError(
				w,
				http.StatusBadRequest,
				"Paid amount cannot be greater than total amount",
				nil,
			)
			return
		}

		// Calculate due
		calculatedDue := calculatedTotal - req.PaidAmount

		// Avoid tiny floating point residue
		if calculatedDue < 0.000001 {
			calculatedDue = 0
		}

		// ======================================
		// Calculate Payment Status
		// ======================================

		var paymentStatus string

		switch {
		case req.PaidAmount == 0:
			paymentStatus = "unpaid"

		case req.PaidAmount < calculatedTotal:
			paymentStatus = "partial"

		default:
			paymentStatus = "paid"
		}

		// ======================================
		// Start Transaction
		// ======================================

		tx, err := db.BeginTx(ctx, nil)

		if err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to start transaction",
				err,
			)
			return
		}

		defer tx.Rollback()

		// ======================================
		// Get Old Purchase Information
		// ======================================

		var (
			oldSupplierID int
			oldDueAmount  float64
		)

		err = tx.QueryRowContext(
			ctx,
			`
			SELECT
				supplier_id,
				due_amount
			FROM purchases
			WHERE id = $1
			`,
			id,
		).Scan(
			&oldSupplierID,
			&oldDueAmount,
		)

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				respondError(
					w,
					http.StatusNotFound,
					"Purchase not found",
					err,
				)
				return
			}

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to get old purchase data",
				err,
			)
			return
		}

		// ======================================
		// Check New Supplier Exists
		// ======================================

		var supplierExists bool

		err = tx.QueryRowContext(
			ctx,
			`
			SELECT EXISTS(
				SELECT 1
				FROM suppliers
				WHERE id = $1
			)
			`,
			req.SupplierID,
		).Scan(&supplierExists)

		if err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to verify supplier",
				err,
			)
			return
		}

		if !supplierExists {
			respondError(
				w,
				http.StatusBadRequest,
				"Supplier not found",
				nil,
			)
			return
		}

		// ======================================
		// Read Old Purchase Items
		// ======================================

		rows, err := tx.QueryContext(
			ctx,
			`
			SELECT
				medicine_id,
				quantity
			FROM purchase_items
			WHERE purchase_id = $1
			`,
			id,
		)

		if err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to fetch old purchase items",
				err,
			)
			return
		}

		type oldPurchaseItem struct {
			MedicineID int
			Quantity   int
		}

		var oldItems []oldPurchaseItem

		for rows.Next() {

			var item oldPurchaseItem

			if err := rows.Scan(
				&item.MedicineID,
				&item.Quantity,
			); err != nil {

				rows.Close()

				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to read old purchase items",
					err,
				)
				return
			}

			oldItems = append(oldItems, item)
		}

		if err := rows.Err(); err != nil {

			rows.Close()

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed while reading old purchase items",
				err,
			)
			return
		}

		rows.Close()

		// ======================================
		// Rollback Old Medicine Stock
		// ======================================

		for _, item := range oldItems {

			var currentStock int

			err = tx.QueryRowContext(
				ctx,
				`
				SELECT
					current_stock
				FROM medicines
				WHERE id = $1
				FOR UPDATE
				`,
				item.MedicineID,
			).Scan(&currentStock)

			if err != nil {

				if errors.Is(err, sql.ErrNoRows) {
					respondError(
						w,
						http.StatusBadRequest,
						fmt.Sprintf(
							"Medicine %d from old purchase no longer exists",
							item.MedicineID,
						),
						err,
					)
					return
				}

				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to check medicine stock",
					err,
				)
				return
			}

			// Prevent negative stock
			if currentStock < item.Quantity {

				respondError(
					w,
					http.StatusConflict,
					fmt.Sprintf(
						"Cannot update purchase because medicine ID %d has insufficient current stock to rollback the old quantity",
						item.MedicineID,
					),
					nil,
				)
				return
			}

			_, err = tx.ExecContext(
				ctx,
				`
				UPDATE medicines
				SET
					current_stock = current_stock - $1,
					updated_at = NOW()
				WHERE id = $2
				`,
				item.Quantity,
				item.MedicineID,
			)

			if err != nil {
				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to rollback medicine stock",
					err,
				)
				return
			}
		}

		// ======================================
		// Rollback Old Supplier Balance
		// ======================================

		_, err = tx.ExecContext(
			ctx,
			`
			UPDATE suppliers
			SET
				current_balance = current_balance - $1,
				updated_at = NOW()
			WHERE id = $2
			`,
			oldDueAmount,
			oldSupplierID,
		)

		if err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to rollback supplier balance",
				err,
			)
			return
		}

		// ======================================
		// Delete Old Purchase Items
		// ======================================

		_, err = tx.ExecContext(
			ctx,
			`
			DELETE FROM purchase_items
			WHERE purchase_id = $1
			`,
			id,
		)

		if err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to delete old purchase items",
				err,
			)
			return
		}

		// ======================================
		// Update Purchase Header
		// ======================================

		_, err = tx.ExecContext(
			ctx,
			`
			UPDATE purchases
			SET
				invoice_no = $1,
				supplier_id = $2,
				purchase_date = $3,
				subtotal = $4,
				discount = $5,
				tax = $6,
				total_amount = $7,
				paid_amount = $8,
				due_amount = $9,
				payment_status = $10,
				note = $11,
				updated_at = NOW()
			WHERE id = $12
			`,
			req.InvoiceNo,
			req.SupplierID,
			req.PurchaseDate,
			calculatedSubTotal,
			req.Discount,
			req.Tax,
			calculatedTotal,
			req.PaidAmount,
			calculatedDue,
			paymentStatus,
			req.Note,
			id,
		)

		if err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to update purchase",
				err,
			)
			return
		}

		// ======================================
		// Insert New Purchase Items
		// ======================================

		// ======================================
		// Insert New Purchase Items
		// ======================================

		for _, item := range req.Items {

			totalPrice := float64(item.Quantity) * item.PurchasePrice

			_, err = tx.ExecContext(
				ctx,
				`
        INSERT INTO purchase_items
        (
            purchase_id,
            medicine_id,
            batch_no,
            expiry_date,
            quantity,
            purchase_price,
            selling_price,
            total_price,
            created_at
        )
        VALUES
        (
            $1,
            $2,
            $3,
            $4,
            $5,
            $6,
            $7,
            $8,
            NOW()
        )
        `,
				id,
				item.MedicineID,
				item.BatchNo,
				item.ExpiryDate,
				item.Quantity,
				item.PurchasePrice,
				item.SalePrice,
				totalPrice,
			)

			if err != nil {
				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to insert purchase items",
					err,
				)
				return
			}

			// ======================================
			// Increase New Medicine Stock
			// ======================================

			_, err = tx.ExecContext(
				ctx,
				`
        UPDATE medicines
        SET
            current_stock = current_stock + $1,
            selling_price = $2,
            mrp = $3,
            updated_at = NOW()
        WHERE id = $4
        `,
				item.Quantity,
				item.SalePrice,
				item.SalePrice,
				item.MedicineID,
			)

			if err != nil {
				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to update medicine stock",
					err,
				)
				return
			}
		}

		// ======================================
		// Update New Supplier Balance
		// ======================================

		_, err = tx.ExecContext(
			ctx,
			`
			UPDATE suppliers
			SET
				current_balance = current_balance + $1,
				updated_at = NOW()
			WHERE id = $2
			`,
			calculatedDue,
			req.SupplierID,
		)

		if err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to update supplier balance",
				err,
			)
			return
		}

		// ======================================
		// Commit Transaction
		// ======================================

		if err := tx.Commit(); err != nil {
			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to complete update purchase",
				err,
			)
			return
		}

		// ======================================
		// Success Response
		// ======================================

		respondSuccess(
			w,
			http.StatusOK,
			"Purchase updated successfully",
			map[string]interface{}{
				"purchaseID":    id,
				"invoiceNo":     req.InvoiceNo,
				"subTotal":      calculatedSubTotal,
				"totalAmount":   calculatedTotal,
				"paidAmount":    req.PaidAmount,
				"dueAmount":     calculatedDue,
				"paymentStatus": paymentStatus,
			},
		)
	}
}

func DeletePurchase(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		purchaseID, err := parseIDFromURL(r)

		if err != nil {

			respondError(
				w,
				http.StatusBadRequest,
				"Invalid purchase id",
				err,
			)

			return
		}

		tx, err := db.BeginTx(ctx, nil)

		if err != nil {

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to start transaction",
				err,
			)

			return
		}

		defer tx.Rollback()

		/*
			Get Purchase Information
		*/

		var (
			supplierID int
			dueAmount  float64
			status     int
		)

		err = tx.QueryRowContext(
			ctx,

			`
			SELECT
				supplier_id,
				due_amount,
				status

			FROM purchases

			WHERE id=$1
			`,

			purchaseID,
		).Scan(

			&supplierID,
			&dueAmount,
			&status,
		)

		if errors.Is(err, sql.ErrNoRows) {

			respondError(
				w,
				http.StatusNotFound,
				"Purchase not found",
				err,
			)

			return
		}

		if err != nil {

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to fetch purchase",
				err,
			)

			return
		}

		/*
			Check Already Deleted
		*/

		if status == 0 {

			respondError(
				w,
				http.StatusBadRequest,
				"Purchase already deleted",
				errors.New("purchase already deleted"),
			)

			return
		}

		/*
			Get Purchase Items
		*/

		rows, err := tx.QueryContext(

			ctx,

			`
			SELECT
				medicine_id,
				quantity

			FROM purchase_items

			WHERE purchase_id=$1
			`,

			purchaseID,
		)

		if err != nil {

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to fetch purchase items",
				err,
			)

			return
		}

		for rows.Next() {

			var (
				medicineID int
				quantity   int
			)

			err = rows.Scan(
				&medicineID,
				&quantity,
			)

			if err != nil {

				rows.Close()

				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to scan purchase items",
					err,
				)

				return
			}

			/*
				Reverse Medicine Stock
			*/

			_, err = tx.ExecContext(

				ctx,

				`
				UPDATE medicines

				SET
					current_stock = current_stock - $1,
					updated_at = NOW()

				WHERE id=$2
				`,

				quantity,
				medicineID,
			)

			if err != nil {

				rows.Close()

				respondError(
					w,
					http.StatusInternalServerError,
					"Failed to reverse medicine stock",
					err,
				)

				return
			}

		}

		rows.Close()

		if err := rows.Err(); err != nil {

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to process purchase items",
				err,
			)

			return
		}

		/*
			Reverse Supplier Balance
		*/

		_, err = tx.ExecContext(

			ctx,

			`
			UPDATE suppliers

			SET
				current_balance = current_balance - $1,
				updated_at = NOW()

			WHERE id=$2
			`,

			dueAmount,
			supplierID,
		)

		if err != nil {

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to reverse supplier balance",
				err,
			)

			return
		}

		/*
			Soft Delete Purchase Items
		*/

		_, err = tx.ExecContext(

			ctx,

			`
			UPDATE purchase_items

			SET
				status = 0,
				updated_at = NOW()

			WHERE purchase_id=$1
			`,

			purchaseID,
		)

		if err != nil {

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to delete purchase items",
				err,
			)

			return
		}

		/*
			Soft Delete Purchase
		*/

		_, err = tx.ExecContext(

			ctx,

			`
			UPDATE purchases

			SET
				status = 0,
				updated_at = NOW()

			WHERE id=$1
			`,

			purchaseID,
		)

		if err != nil {

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to delete purchase",
				err,
			)

			return
		}

		/*
			Commit Transaction
		*/

		if err := tx.Commit(); err != nil {

			log.Println(
				"Delete Purchase Commit:",
				err,
			)

			respondError(
				w,
				http.StatusInternalServerError,
				"Failed to complete delete purchase",
				err,
			)

			return
		}

		respondSuccess(

			w,

			http.StatusOK,

			"Purchase deleted successfully",

			map[string]interface{}{

				"purchaseID": purchaseID,
			},
		)

	}
}
