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
	UserID        int                   `json:"userID"`
	InvoiceNo     string                `json:"invoiceNo"`
	SupplierID    int                   `json:"supplierID"`
	PurchaseDate  string                `json:"purchaseDate"`
	SubTotal      float64               `json:"subTotal"`
	Discount      float64               `json:"discount"`
	Tax           float64               `json:"tax"`
	TotalAmount   float64               `json:"totalAmount"`
	PaidAmount    float64               `json:"paidAmount"`
	DueAmount     float64               `json:"dueAmount"`
	PaymentStatus string                `json:"paymentStatus"`
	Note          string                `json:"note"`
	Items         []PurchaseItemRequest `json:"items"`
}

type DeletePurchaseRequest struct {
	Reason string `json:"reason"`
	UserID int    `json:"userID"`
}

type ReturnItemRequest struct {
	PurchaseItemID int    `json:"purchaseItemID"`
	MedicineID     int    `json:"medicineID"`
	ReturnQty      int    `json:"returnQty"`
	ItemReason     string `json:"itemReason"`
}

type CreatePurchaseReturnRequest struct {
	UserID       int                 `json:"userID"`
	PurchaseID   int                 `json:"purchaseID"`
	ReturnType   string              `json:"returnType"`
	ReturnReason string              `json:"returnReason"`
	Items        []ReturnItemRequest `json:"items,omitempty"`
}

func respondJSON(w http.ResponseWriter, statusCode int, res Response) {
	res.Code = statusCode
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("Failed to encode response:", err)
	}
}

func respondError(w http.ResponseWriter, statusCode int, message string, err error) {
	res := Response{
		Status:  false,
		Message: message,
	}
	if err != nil {
		res.Error = err.Error()
	}
	respondJSON(w, statusCode, res)
}

func respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
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

// Helper to get logged-in user ID (First preference: Request Body, Second: Context, Third: Fallback 1)
func resolveUserID(reqUserID int, r *http.Request) int {
	if reqUserID > 0 {
		return reqUserID
	}
	if userID, ok := r.Context().Value("user_id").(int); ok && userID > 0 {
		return userID
	}
	return 1
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
	}
	return nil
}

// CreatePurchase Handler
func CreatePurchase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req CreatePurchaseRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request payload", err)
			return
		}

		userID := resolveUserID(req.UserID, r)
		req.InvoiceNo = strings.TrimSpace(req.InvoiceNo)
		req.Note = strings.TrimSpace(req.Note)

		if err := validatePurchase(req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error(), err)
			return
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to start transaction", err)
			return
		}
		defer tx.Rollback()

		var purchaseID int
		purchaseQuery := `
			INSERT INTO purchases (
				invoice_no, supplier_id, user_id, purchase_date, subtotal, discount, tax, 
				total_amount, paid_amount, due_amount, payment_status, note, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 1, NOW(), NOW())
			RETURNING id
		`
		err = tx.QueryRowContext(ctx, purchaseQuery,
			req.InvoiceNo, req.SupplierID, userID, req.PurchaseDate, req.SubTotal, req.Discount, req.Tax,
			req.TotalAmount, req.PaidAmount, req.DueAmount, req.PaymentStatus, req.Note,
		).Scan(&purchaseID)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create purchase", err)
			return
		}

		itemQuery := `
			INSERT INTO purchase_items (
				purchase_id, medicine_id, batch_no, expiry_date, quantity, purchase_price, selling_price, total_price, created_at
			) VALUES ($1, $2, $3, NULLIF($4, '')::DATE, $5, $6, $7, $8, NOW())
		`
		updateMedicineQuery := `
			UPDATE medicines SET current_stock = current_stock + $1, selling_price = $2, mrp = $3, updated_at = NOW() WHERE id = $4
		`

		for _, item := range req.Items {
			totalPrice := float64(item.Quantity) * item.PurchasePrice
			_, err = tx.ExecContext(ctx, itemQuery, purchaseID, item.MedicineID, item.BatchNo, item.ExpiryDate, item.Quantity, item.PurchasePrice, item.SalePrice, totalPrice)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to save purchase items", err)
				return
			}

			_, err = tx.ExecContext(ctx, updateMedicineQuery, item.Quantity, item.SalePrice, item.SalePrice, item.MedicineID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to update medicine stock", err)
				return
			}
		}

		_, err = tx.ExecContext(ctx, `UPDATE suppliers SET current_balance = current_balance + $1, updated_at = NOW() WHERE id = $2`, req.DueAmount, req.SupplierID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update supplier balance", err)
			return
		}

		if err := tx.Commit(); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to complete purchase", err)
			return
		}

		respondSuccess(w, http.StatusCreated, "Purchase created successfully", map[string]interface{}{"purchaseID": purchaseID, "userID": userID})
	}
}

// UpdatePurchase Handler
// UpdatePurchase Handler
func UpdatePurchase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := parseIDFromURL(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid purchase ID", err)
			return
		}

		var req CreatePurchaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request payload", err)
			return
		}

		userID := resolveUserID(req.UserID, r)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to start transaction", err)
			return
		}
		defer tx.Rollback()

		var oldSupplierID int
		var oldDueAmount float64
		err = tx.QueryRowContext(ctx, `SELECT supplier_id, due_amount FROM purchases WHERE id = $1 AND status = 1`, id).Scan(&oldSupplierID, &oldDueAmount)
		if err != nil {
			respondError(w, http.StatusNotFound, "Purchase not found or already deleted", err)
			return
		}

		// ১. পুরোনো আইটেমগুলো ফেচ করে একটি স্লাইস/ম্যাপে সংরক্ষণ (Transaction Lock ও Protocol Mismatch এড়াতে)
		type oldItem struct {
			medicineID int
			quantity   int
		}
		var oldItems []oldItem

		rows, err := tx.QueryContext(ctx, `SELECT medicine_id, quantity FROM purchase_items WHERE purchase_id = $1`, id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to fetch old items", err)
			return
		}

		for rows.Next() {
			var item oldItem
			if err := rows.Scan(&item.medicineID, &item.quantity); err != nil {
				rows.Close()
				respondError(w, http.StatusInternalServerError, "Scan error", err)
				return
			}
			oldItems = append(oldItems, item)
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			respondError(w, http.StatusInternalServerError, "Failed to iterate old items", err)
			return
		}
		rows.Close() // Rows বাধ্যতামূলকভাবে বন্ধ করতে হবে

		// ২. পুরোনো স্টোক রোলব্যাক (Stock Rollback)
		for _, item := range oldItems {
			_, err = tx.ExecContext(ctx, `
				UPDATE medicines 
				SET current_stock = GREATEST(0, current_stock - $1), updated_at = NOW() 
				WHERE id = $2
			`, item.quantity, item.medicineID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to rollback stock", err)
				return
			}
		}

		// ৩. পুরোনো সাপ্লায়ার ব্যালেন্স রোলব্যাক
		_, err = tx.ExecContext(ctx, `
			UPDATE suppliers 
			SET current_balance = GREATEST(0, current_balance - $1), updated_at = NOW() 
			WHERE id = $2
		`, oldDueAmount, oldSupplierID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to rollback supplier balance", err)
			return
		}

		// ৪. পুরোনো আইটেমগুলো ডিলিট করা
		_, err = tx.ExecContext(ctx, `DELETE FROM purchase_items WHERE purchase_id = $1`, id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to clear old items", err)
			return
		}

		// ৫. নতুন আইটেম যুক্ত করা ও স্টক আপডেট করা
		for _, item := range req.Items {
			totalPrice := float64(item.Quantity) * item.PurchasePrice
			_, err = tx.ExecContext(ctx, `
				INSERT INTO purchase_items (
					purchase_id, medicine_id, batch_no, expiry_date, quantity, purchase_price, selling_price, total_price, created_at
				) VALUES ($1, $2, $3, NULLIF($4, '')::DATE, $5, $6, $7, $8, NOW())
			`, id, item.MedicineID, item.BatchNo, item.ExpiryDate, item.Quantity, item.PurchasePrice, item.SalePrice, totalPrice)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to insert new purchase items", err)
				return
			}

			_, err = tx.ExecContext(ctx, `
				UPDATE medicines 
				SET current_stock = current_stock + $1, selling_price = $2, mrp = $3, updated_at = NOW() 
				WHERE id = $4
			`, item.Quantity, item.SalePrice, item.SalePrice, item.MedicineID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to update stock", err)
				return
			}
		}

		// ৬. পারচেস হেডার ও অডিট ট্রেইল আপডেট করা
		_, err = tx.ExecContext(ctx, `
			UPDATE purchases SET 
				invoice_no = $1, supplier_id = $2, purchase_date = $3, subtotal = $4, discount = $5, tax = $6,
				total_amount = $7, paid_amount = $8, due_amount = $9, payment_status = $10, note = $11, 
				updated_by = $12, updated_at = NOW()
			WHERE id = $13
		`, req.InvoiceNo, req.SupplierID, req.PurchaseDate, req.SubTotal, req.Discount, req.Tax, req.TotalAmount, req.PaidAmount, req.DueAmount, req.PaymentStatus, req.Note, userID, id)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update purchase header", err)
			return
		}

		// ৭. নতুন সাপ্লায়ার ব্যালেন্স আপডেট করা
		_, err = tx.ExecContext(ctx, `
			UPDATE suppliers 
			SET current_balance = current_balance + $1, updated_at = NOW() 
			WHERE id = $2
		`, req.DueAmount, req.SupplierID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update new supplier balance", err)
			return
		}

		// ৮. ট্রানজেকশন কমিক করা
		if err := tx.Commit(); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to commit update", err)
			return
		}

		respondSuccess(w, http.StatusOK, "Purchase updated successfully", map[string]interface{}{
			"purchaseID": id,
			"updatedBy":  userID,
		})
	}
}

// DeletePurchase Handler
func DeletePurchase(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		purchaseID, err := parseIDFromURL(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid purchase ID", err)
			return
		}

		var req DeletePurchaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
			respondError(w, http.StatusBadRequest, "Delete reason is required", err)
			return
		}

		userID := resolveUserID(req.UserID, r)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to start transaction", err)
			return
		}
		defer tx.Rollback()

		var supplierID int
		var dueAmount float64
		var status int
		err = tx.QueryRowContext(ctx, `SELECT supplier_id, due_amount, status FROM purchases WHERE id = $1`, purchaseID).Scan(&supplierID, &dueAmount, &status)
		if errors.Is(err, sql.ErrNoRows) || status == 0 {
			respondError(w, http.StatusNotFound, "Purchase not found or already deleted", err)
			return
		}

		rows, err := tx.QueryContext(ctx, `SELECT medicine_id, quantity FROM purchase_items WHERE purchase_id = $1`, purchaseID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to fetch items", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var medID, qty int
			if err := rows.Scan(&medID, &qty); err != nil {
				respondError(w, http.StatusInternalServerError, "Scan error", err)
				return
			}
			_, err = tx.ExecContext(ctx, `UPDATE medicines SET current_stock = GREATEST(0, current_stock - $1), updated_at = NOW() WHERE id = $2`, qty, medID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to reverse stock", err)
				return
			}
		}
		if err := rows.Err(); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to read purchase items", err)
			return
		}

		_, err = tx.ExecContext(ctx, `UPDATE suppliers SET current_balance = GREATEST(0, current_balance - $1), updated_at = NOW() WHERE id = $2`, dueAmount, supplierID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to reverse supplier balance", err)
			return
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE purchases 
			SET status = 0, deleted_by = $1, deleted_at = NOW(), delete_reason = $2, updated_at = NOW()
			WHERE id = $3
		`, userID, req.Reason, purchaseID)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to delete purchase", err)
			return
		}

		if err := tx.Commit(); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to complete deletion", err)
			return
		}

		respondSuccess(w, http.StatusOK, "Purchase deleted successfully", map[string]interface{}{
			"purchaseID": purchaseID,
			"deletedBy":  userID,
			"deletedAt":  time.Now(),
		})
	}
}

// ProcessPurchaseReturn Handler
func ProcessPurchaseReturn(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req CreatePurchaseReturnRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request payload", err)
			return
		}

		userID := resolveUserID(req.UserID, r)
		req.ReturnType = strings.ToLower(strings.TrimSpace(req.ReturnType))
		req.ReturnReason = strings.TrimSpace(req.ReturnReason)

		if req.PurchaseID <= 0 {
			respondError(w, http.StatusBadRequest, "Purchase ID is required", nil)
			return
		}
		if req.ReturnType != "invoice_wise" && req.ReturnType != "item_wise" {
			respondError(w, http.StatusBadRequest, "Invalid returnType. Allowed: 'invoice_wise' or 'item_wise'", nil)
			return
		}
		if req.ReturnReason == "" {
			respondError(w, http.StatusBadRequest, "Return reason is required", nil)
			return
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to start transaction", err)
			return
		}
		defer tx.Rollback()

		var supplierID int
		var dueAmount, totalAmount float64
		var status int

		err = tx.QueryRowContext(ctx, `SELECT supplier_id, due_amount, total_amount, status FROM purchases WHERE id = $1`, req.PurchaseID).Scan(&supplierID, &dueAmount, &totalAmount, &status)
		if errors.Is(err, sql.ErrNoRows) || status == 0 {
			respondError(w, http.StatusNotFound, "Purchase invoice not found or deleted", err)
			return
		}

		var totalReturnAmount float64

		var returnID int
		err = tx.QueryRowContext(ctx, `
			INSERT INTO purchase_returns (purchase_id, supplier_id, user_id, return_type, total_return_amount, return_reason, created_at)
			VALUES ($1, $2, $3, $4, 0.00, $5, NOW()) RETURNING id
		`, req.PurchaseID, supplierID, userID, req.ReturnType, req.ReturnReason).Scan(&returnID)

		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to initialize purchase return", err)
			return
		}

		if req.ReturnType == "invoice_wise" {
			rows, err := tx.QueryContext(ctx, `SELECT id, medicine_id, quantity, purchase_price FROM purchase_items WHERE purchase_id = $1`, req.PurchaseID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to fetch purchase items", err)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var pItemID, medID, qty int
				var price float64
				if err := rows.Scan(&pItemID, &medID, &qty, &price); err != nil {
					respondError(w, http.StatusInternalServerError, "Scan error", err)
					return
				}
				itemTotal := float64(qty) * price
				totalReturnAmount += itemTotal

				_, err = tx.ExecContext(ctx, `
					INSERT INTO purchase_return_items (purchase_return_id, purchase_item_id, medicine_id, return_qty, unit_price, total_price, item_reason, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
				`, returnID, pItemID, medID, qty, price, itemTotal, req.ReturnReason)
				if err != nil {
					respondError(w, http.StatusInternalServerError, "Failed to insert return item details", err)
					return
				}

				_, err = tx.ExecContext(ctx, `UPDATE medicines SET current_stock = GREATEST(0, current_stock - $1), updated_at = NOW() WHERE id = $2`, qty, medID)
				if err != nil {
					respondError(w, http.StatusInternalServerError, "Failed to adjust medicine stock", err)
					return
				}
			}
			if err = rows.Err(); err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to iterate purchase items", err)
				return
			}
		} else {
			if len(req.Items) == 0 {
				respondError(w, http.StatusBadRequest, "At least one item must be selected for item-wise return", nil)
				return
			}

			for _, retItem := range req.Items {
				if retItem.ReturnQty <= 0 {
					continue
				}

				var origQty int
				var purchasePrice float64
				err := tx.QueryRowContext(ctx, `SELECT quantity, purchase_price FROM purchase_items WHERE id = $1 AND purchase_id = $2`, retItem.PurchaseItemID, req.PurchaseID).Scan(&origQty, &purchasePrice)
				if err != nil {
					respondError(w, http.StatusBadRequest, fmt.Sprintf("Item ID %d not found in original purchase", retItem.PurchaseItemID), err)
					return
				}

				if retItem.ReturnQty > origQty {
					respondError(w, http.StatusBadRequest, fmt.Sprintf("Return Qty (%d) cannot exceed purchased Qty (%d)", retItem.ReturnQty, origQty), nil)
					return
				}

				itemTotal := float64(retItem.ReturnQty) * purchasePrice
				totalReturnAmount += itemTotal

				itemReason := strings.TrimSpace(retItem.ItemReason)
				if itemReason == "" {
					itemReason = req.ReturnReason
				}

				_, err = tx.ExecContext(ctx, `
					INSERT INTO purchase_return_items (purchase_return_id, purchase_item_id, medicine_id, return_qty, unit_price, total_price, item_reason, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
				`, returnID, retItem.PurchaseItemID, retItem.MedicineID, retItem.ReturnQty, purchasePrice, itemTotal, itemReason)

				if err != nil {
					respondError(w, http.StatusInternalServerError, "Failed to log return item", err)
					return
				}

				_, err = tx.ExecContext(ctx, `UPDATE medicines SET current_stock = GREATEST(0, current_stock - $1), updated_at = NOW() WHERE id = $2`, retItem.ReturnQty, retItem.MedicineID)
				if err != nil {
					respondError(w, http.StatusInternalServerError, "Failed to update stock for returned item", err)
					return
				}
			}
		}

		_, err = tx.ExecContext(ctx, `UPDATE purchase_returns SET total_return_amount = $1 WHERE id = $2`, totalReturnAmount, returnID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to update return total", err)
			return
		}

		_, err = tx.ExecContext(ctx, `UPDATE suppliers SET current_balance = GREATEST(0, current_balance - $1), updated_at = NOW() WHERE id = $2`, totalReturnAmount, supplierID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to adjust supplier balance", err)
			return
		}

		if err := tx.Commit(); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to complete purchase return transaction", err)
			return
		}

		respondSuccess(w, http.StatusOK, "Purchase return processed successfully", map[string]interface{}{
			"returnID":          returnID,
			"purchaseID":        req.PurchaseID,
			"returnType":        req.ReturnType,
			"totalReturnAmount": totalReturnAmount,
			"returnedBy":        userID,
			"returnedAt":        time.Now(),
		})
	}
}

// GetPurchases
func GetPurchases(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userIDParam := r.URL.Query().Get("user_id")

		query := `
			SELECT
				p.id,
				p.invoice_no,
				p.supplier_id,
				COALESCE(s.supplier_name, '') AS supplier_name,
				p.user_id,
				COALESCE(TRIM(u.firstname || ' ' || u.lastname), '') AS created_by_name,
				p.purchase_date,
				p.subtotal,
				p.discount,
				p.tax,
				p.total_amount,
				p.paid_amount,
				p.due_amount,
				p.payment_status,
				p.status,
				p.delete_reason,
				p.deleted_at,
				p.created_at,
				p.updated_at,
				COALESCE((SELECT SUM(total_return_amount) FROM purchase_returns WHERE purchase_id = p.id), 0.00) AS total_returned_amount
			FROM purchases p
			LEFT JOIN suppliers s ON p.supplier_id = s.id
			LEFT JOIN users u ON p.user_id = u.id
		`

		var args []interface{}
		if userIDParam != "" {
			query += " WHERE p.user_id = $1 "
			args = append(args, userIDParam)
		}

		query += " ORDER BY p.id DESC"

		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			log.Println("GetPurchases Error:", err)
			respondError(w, http.StatusInternalServerError, "Failed to fetch purchases", err)
			return
		}
		defer rows.Close()

		purchases := make([]map[string]interface{}, 0)

		for rows.Next() {
			var (
				id                  int
				invoiceNo           string
				supplierID          int
				supplierName        string
				userID              int
				createdByName       string
				purchaseDate        time.Time
				subtotal            float64
				discount            float64
				tax                 float64
				totalAmount         float64
				paidAmount          float64
				dueAmount           float64
				paymentStatus       string
				status              int
				deleteReason        sql.NullString
				deletedAt           sql.NullTime
				createdAt           time.Time
				updatedAt           time.Time
				totalReturnedAmount float64
			)

			err := rows.Scan(
				&id,
				&invoiceNo,
				&supplierID,
				&supplierName,
				&userID,
				&createdByName,
				&purchaseDate,
				&subtotal,
				&discount,
				&tax,
				&totalAmount,
				&paidAmount,
				&dueAmount,
				&paymentStatus,
				&status,
				&deleteReason,
				&deletedAt,
				&createdAt,
				&updatedAt,
				&totalReturnedAmount,
			)

			if err != nil {
				log.Println("GetPurchases Scan Error:", err)
				continue
			}

			var formattedDeletedAt interface{} = nil
			if deletedAt.Valid {
				formattedDeletedAt = deletedAt.Time.Format(time.RFC3339)
			}

			purchases = append(purchases, map[string]interface{}{
				"id":                  id,
				"invoiceNo":           invoiceNo,
				"supplierID":          supplierID,
				"supplierName":        supplierName,
				"userID":              userID,
				"createdBy":           createdByName,
				"purchaseDate":        purchaseDate.Format(time.RFC3339),
				"subtotal":            subtotal,
				"discount":            discount,
				"tax":                 tax,
				"totalAmount":         totalAmount,
				"paidAmount":          paidAmount,
				"dueAmount":           dueAmount,
				"paymentStatus":       paymentStatus,
				"status":              status,
				"deleteReason":        deleteReason.String,
				"deletedAt":           formattedDeletedAt,
				"totalReturnedAmount": totalReturnedAmount,
				"createdAt":           createdAt.Format(time.RFC3339),
				"updatedAt":           updatedAt.Format(time.RFC3339),
			})
		}

		if err := rows.Err(); err != nil {
			log.Println("GetPurchases Rows Error:", err)
			respondError(w, http.StatusInternalServerError, "Failed to fetch purchases", err)
			return
		}

		respondSuccess(w, http.StatusOK, "Purchases retrieved successfully", purchases)
	}
}

// GetPurchaseByID Handler
func GetPurchaseByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		id, err := parseIDFromURL(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid purchase ID", err)
			return
		}

		purchaseQuery := `
			SELECT
				p.id,
				p.invoice_no,
				p.supplier_id,
				COALESCE(s.supplier_name, '') AS supplier_name,
				COALESCE(s.mobile, '') AS supplier_mobile,
				COALESCE(s.address, '') AS supplier_address,
				p.user_id,
				COALESCE(TRIM(u1.firstname || ' ' || u1.lastname), '') AS created_by_name,
				p.updated_by,
				COALESCE(TRIM(u2.firstname || ' ' || u2.lastname), '') AS updated_by_name,
				p.deleted_by,
				COALESCE(TRIM(u3.firstname || ' ' || u3.lastname), '') AS deleted_by_name,
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
				p.delete_reason,
				p.deleted_at,
				p.created_at,
				p.updated_at
			FROM purchases p
			LEFT JOIN suppliers s ON p.supplier_id = s.id
			LEFT JOIN users u1 ON p.user_id = u1.id
			LEFT JOIN users u2 ON p.updated_by = u2.id
			LEFT JOIN users u3 ON p.deleted_by = u3.id
			WHERE p.id = $1
		`

		var (
			purchaseID      int
			invoiceNo       string
			supplierID      int
			supplierName    string
			supplierMobile  string
			supplierAddress string
			userID          int
			createdByName   string
			updatedBy       sql.NullInt64
			updatedByName   string
			deletedBy       sql.NullInt64
			deletedByName   string
			purchaseDate    time.Time
			subTotal        float64
			discount        float64
			tax             float64
			totalAmount     float64
			paidAmount      float64
			dueAmount       float64
			paymentStatus   string
			note            sql.NullString
			status          int
			deleteReason    sql.NullString
			deletedAt       sql.NullTime
			createdAt       time.Time
			updatedAt       time.Time
		)

		err = db.QueryRowContext(ctx, purchaseQuery, id).Scan(
			&purchaseID,
			&invoiceNo,
			&supplierID,
			&supplierName,
			&supplierMobile,
			&supplierAddress,
			&userID,
			&createdByName,
			&updatedBy,
			&updatedByName,
			&deletedBy,
			&deletedByName,
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
			&deleteReason,
			&deletedAt,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondError(w, http.StatusNotFound, "Purchase not found", err)
				return
			}
			log.Println("GetPurchaseByID Header Error:", err)
			respondError(w, http.StatusInternalServerError, "Failed to fetch purchase details", err)
			return
		}

		var formattedDeletedAt interface{} = nil
		if deletedAt.Valid {
			formattedDeletedAt = deletedAt.Time.Format(time.RFC3339)
		}

		purchase := map[string]interface{}{
			"id":        purchaseID,
			"invoiceNo": invoiceNo,
			"supplier": map[string]interface{}{
				"id":      supplierID,
				"name":    supplierName,
				"mobile":  supplierMobile,
				"address": supplierAddress,
			},
			"audit": map[string]interface{}{
				"createdBy":    createdByName,
				"createdById":  userID,
				"updatedBy":    updatedByName,
				"updatedById":  updatedBy.Int64,
				"deletedBy":    deletedByName,
				"deletedById":  deletedBy.Int64,
				"deletedAt":    formattedDeletedAt,
				"deleteReason": deleteReason.String,
			},
			"purchaseDate":  purchaseDate.Format(time.RFC3339),
			"subtotal":      subTotal,
			"discount":      discount,
			"tax":           tax,
			"totalAmount":   totalAmount,
			"paidAmount":    paidAmount,
			"dueAmount":     dueAmount,
			"paymentStatus": paymentStatus,
			"note":          note.String,
			"status":        status,
			"createdAt":     createdAt.Format(time.RFC3339),
			"updatedAt":     updatedAt.Format(time.RFC3339),
		}

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
				COALESCE((SELECT SUM(return_qty) FROM purchase_return_items WHERE purchase_item_id = pi.id), 0) AS returned_qty,
				pi.purchase_price,
				pi.selling_price,
				pi.total_price
			FROM purchase_items pi
			INNER JOIN medicines m ON pi.medicine_id = m.id
			WHERE pi.purchase_id = $1
			ORDER BY pi.id ASC
		`

		rows, err := db.QueryContext(ctx, itemQuery, purchaseID)
		if err != nil {
			log.Println("GetPurchaseByID Items Error:", err)
			respondError(w, http.StatusInternalServerError, "Failed to fetch purchase items", err)
			return
		}
		defer rows.Close()

		items := make([]map[string]interface{}, 0)

		for rows.Next() {
			var (
				itemID           int
				medicineID       int
				medicineCode     sql.NullString
				medicineBarcode  sql.NullString
				medicineName     sql.NullString
				medicineStrength sql.NullString
				medicineGeneric  sql.NullString
				batchNo          sql.NullString
				expiryDate       sql.NullTime
				quantity         int
				returnedQty      int
				purchasePrice    float64
				sellingPrice     float64
				totalPrice       float64
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
				&returnedQty,
				&purchasePrice,
				&sellingPrice,
				&totalPrice,
			)

			if err != nil {
				log.Println("GetPurchaseByID Item Scan Error:", err)
				respondError(w, http.StatusInternalServerError, "Failed to read purchase item", err)
				return
			}

			var formattedExpiryDate interface{} = nil
			if expiryDate.Valid {
				formattedExpiryDate = expiryDate.Time.Format(time.RFC3339)
			}

			items = append(items, map[string]interface{}{
				"id":         itemID,
				"medicineID": medicineID,
				"medicine": map[string]interface{}{
					"code":     medicineCode.String,
					"barcode":  medicineBarcode.String,
					"name":     medicineName.String,
					"strength": medicineStrength.String,
					"generic":  medicineGeneric.String,
				},
				"batchNo":       batchNo.String,
				"expiryDate":    formattedExpiryDate,
				"quantity":      quantity,
				"returnedQty":   returnedQty,
				"availableQty":  quantity - returnedQty,
				"purchasePrice": purchasePrice,
				"sellingPrice":  sellingPrice,
				"totalPrice":    totalPrice,
			})
		}

		if err := rows.Err(); err != nil {
			log.Println("GetPurchaseByID Items Iteration Error:", err)
			respondError(w, http.StatusInternalServerError, "Failed to iterate purchase items", err)
			return
		}

		purchase["items"] = items
		respondSuccess(w, http.StatusOK, "Purchase details retrieved successfully", purchase)
	}
}
