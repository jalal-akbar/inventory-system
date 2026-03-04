package repository

import (
	"database/sql"
	"inventory-system/internal/domain"
)

type ProductBatchRepository interface {
	FindByProduct(productId int) ([]domain.ProductBatch, error)
	VerifyByProduct(productId int) error
	Verify(id int) error
	UpdateStock(id, quantity int) error
	GetPendingCount() (int, error)
	Create(b *domain.ProductBatch) (int, error)
	GetWithProduct(id int) (map[string]interface{}, error)
	GetAvailableForProduct(productId int) ([]domain.ProductBatch, error)
	SetStock(id, stock int) error
	GetExpiringCount(days int) (int, error)
	GetExpiredCount() (int, error)
	GetExpiringBatches(days int) ([]map[string]interface{}, error)
	FindByID(id int) (*domain.ProductBatch, error)
	GetProductStatus(productID int) (string, error)
	WithTx(tx *sql.Tx) ProductBatchRepository
}

type mysqlBatchRepository struct {
	db         DBExecutor // Use DBExecutor interface
	originalDB *sql.DB    // Keep original for WithTx
}

// We need to define Executor inside the repo or use the one I created.
// To avoid circular dependency or import issues if I just created executor.go in same package,
// I'll just use the local package interface.

func NewBatchRepository(db *sql.DB) ProductBatchRepository {
	return &mysqlBatchRepository{db: db, originalDB: db}
}

func (r *mysqlBatchRepository) getDB() DBExecutor {
	return r.db
}

func (r *mysqlBatchRepository) WithTx(tx *sql.Tx) ProductBatchRepository {
	return &mysqlBatchRepository{db: tx, originalDB: r.originalDB}
}

func (r *mysqlBatchRepository) FindByProduct(productId int) ([]domain.ProductBatch, error) {
	rows, err := r.getDB().Query("SELECT id, product_id, batch_number, expiry_date, initial_qty, current_stock, purchase_price, selling_price, is_verified, created_at FROM product_batches WHERE product_id = ? ORDER BY expiry_date ASC", productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []domain.ProductBatch
	for rows.Next() {
		var b domain.ProductBatch
		if err := rows.Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.InitialQty, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified, &b.CreatedAt); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, nil
}

func (r *mysqlBatchRepository) VerifyByProduct(productId int) error {
	_, err := r.getDB().Exec("UPDATE product_batches SET is_verified = 1 WHERE product_id = ?", productId)
	return err
}

func (r *mysqlBatchRepository) Verify(id int) error {
	_, err := r.getDB().Exec("UPDATE product_batches SET is_verified = 1 WHERE id = ?", id)
	return err
}

func (r *mysqlBatchRepository) UpdateStock(id, quantity int) error {
	_, err := r.getDB().Exec("UPDATE product_batches SET current_stock = current_stock + ? WHERE id = ?", quantity, id)
	return err
}

func (r *mysqlBatchRepository) GetPendingCount() (int, error) {
	var count int
	err := r.getDB().QueryRow("SELECT COUNT(*) FROM product_batches b JOIN products p ON b.product_id = p.id WHERE b.is_verified = 0 AND p.status = 'active'").Scan(&count)
	return count, err
}

func (r *mysqlBatchRepository) Create(b *domain.ProductBatch) (int, error) {
	res, err := r.getDB().Exec("INSERT INTO product_batches (product_id, batch_number, expiry_date, initial_qty, current_stock, purchase_price, selling_price, is_verified, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		b.ProductID, b.BatchNumber, b.ExpiryDate, b.InitialQty, b.CurrentStock, b.PurchasePrice, b.SellingPrice, b.IsVerified)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *mysqlBatchRepository) GetWithProduct(id int) (map[string]interface{}, error) {
	query := "SELECT b.*, p.name as product_name, p.status as product_status FROM product_batches b JOIN products p ON b.product_id = p.id WHERE b.id = ?"
	row := r.getDB().QueryRow(query, id)

	var b domain.ProductBatch
	var productName, productStatus string
	err := row.Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.InitialQty, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified, &b.CreatedAt, &productName, &productStatus)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"batch":          b,
		"product_name":   productName,
		"product_status": productStatus,
	}, nil
}

func (r *mysqlBatchRepository) GetAvailableForProduct(productId int) ([]domain.ProductBatch, error) {
	rows, err := r.getDB().Query("SELECT id, product_id, batch_number, expiry_date, initial_qty, current_stock, purchase_price, selling_price, is_verified FROM product_batches WHERE product_id = ? AND current_stock > 0 AND is_verified = 1 ORDER BY expiry_date ASC", productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []domain.ProductBatch
	for rows.Next() {
		var b domain.ProductBatch
		if err := rows.Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.InitialQty, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, nil
}

func (r *mysqlBatchRepository) SetStock(id, stock int) error {
	_, err := r.getDB().Exec("UPDATE product_batches SET current_stock = ? WHERE id = ?", stock, id)
	return err
}

func (r *mysqlBatchRepository) GetExpiringCount(days int) (int, error) {
	var count int
	err := r.getDB().QueryRow("SELECT COUNT(b.id) FROM product_batches b JOIN products p ON b.product_id = p.id WHERE p.status = 'active' AND p.is_verified = 1 AND b.is_verified = 1 AND b.expiry_date <= date('now', 'localtime', '+' || ? || ' days') AND b.expiry_date >= date('now', 'localtime') AND b.current_stock > 0", days).Scan(&count)
	return count, err
}

func (r *mysqlBatchRepository) GetExpiredCount() (int, error) {
	var count int
	err := r.getDB().QueryRow("SELECT COUNT(b.id) FROM product_batches b JOIN products p ON b.product_id = p.id WHERE p.status = 'active' AND p.is_verified = 1 AND b.is_verified = 1 AND b.expiry_date < date('now', 'localtime') AND b.current_stock > 0").Scan(&count)
	return count, err
}

func (r *mysqlBatchRepository) GetExpiringBatches(days int) ([]map[string]interface{}, error) {
	query := `
		SELECT b.*, p.name as product_name 
		FROM product_batches b 
		JOIN products p ON b.product_id = p.id 
		WHERE p.status = 'active' AND p.is_verified = 1 AND b.is_verified = 1
		AND b.expiry_date <= date('now', 'localtime', '+' || ? || ' days') 
		AND b.expiry_date >= date('now', 'localtime') 
		AND b.current_stock > 0
		ORDER BY b.expiry_date ASC
	`
	rows, err := r.getDB().Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var b domain.ProductBatch
		var productName string
		err := rows.Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.InitialQty, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified, &b.CreatedAt, &productName)
		if err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"batch":        b,
			"product_name": productName,
		})
	}
	return results, nil
}

func (r *mysqlBatchRepository) FindByID(id int) (*domain.ProductBatch, error) {
	b := &domain.ProductBatch{}
	err := r.getDB().QueryRow("SELECT id, product_id, batch_number, expiry_date, initial_qty, current_stock, purchase_price, selling_price, is_verified, created_at FROM product_batches WHERE id = ?", id).
		Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.InitialQty, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified, &b.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func (r *mysqlBatchRepository) GetProductStatus(productID int) (string, error) {
	var status string
	err := r.getDB().QueryRow("SELECT status FROM products WHERE id = ?", productID).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}
