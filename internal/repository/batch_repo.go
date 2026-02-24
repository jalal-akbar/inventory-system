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
}

type mysqlBatchRepository struct {
	db *sql.DB
}

func NewBatchRepository(db *sql.DB) ProductBatchRepository {
	return &mysqlBatchRepository{db: db}
}

func (r *mysqlBatchRepository) FindByProduct(productId int) ([]domain.ProductBatch, error) {
	rows, err := r.db.Query("SELECT id, product_id, batch_number, expiry_date, current_stock, purchase_price, selling_price, is_verified, created_at FROM product_batches WHERE product_id = ? ORDER BY expiry_date ASC", productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []domain.ProductBatch
	for rows.Next() {
		var b domain.ProductBatch
		if err := rows.Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified, &b.CreatedAt); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, nil
}

func (r *mysqlBatchRepository) VerifyByProduct(productId int) error {
	_, err := r.db.Exec("UPDATE product_batches SET is_verified = 1 WHERE product_id = ?", productId)
	return err
}

func (r *mysqlBatchRepository) Verify(id int) error {
	_, err := r.db.Exec("UPDATE product_batches SET is_verified = 1 WHERE id = ?", id)
	return err
}

func (r *mysqlBatchRepository) UpdateStock(id, quantity int) error {
	_, err := r.db.Exec("UPDATE product_batches SET current_stock = current_stock + ? WHERE id = ?", quantity, id)
	return err
}

func (r *mysqlBatchRepository) GetPendingCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM product_batches b JOIN products p ON b.product_id = p.id WHERE b.is_verified = 0 AND p.status = 'active'").Scan(&count)
	return count, err
}

func (r *mysqlBatchRepository) Create(b *domain.ProductBatch) (int, error) {
	res, err := r.db.Exec("INSERT INTO product_batches (product_id, batch_number, expiry_date, current_stock, purchase_price, selling_price, is_verified, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		b.ProductID, b.BatchNumber, b.ExpiryDate, b.CurrentStock, b.PurchasePrice, b.SellingPrice, b.IsVerified)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *mysqlBatchRepository) GetWithProduct(id int) (map[string]interface{}, error) {
	query := "SELECT b.*, p.name as product_name FROM product_batches b JOIN products p ON b.product_id = p.id WHERE b.id = ?"
	row := r.db.QueryRow(query, id)

	var b domain.ProductBatch
	var productName string
	err := row.Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified, &b.CreatedAt, &productName)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"batch":        b,
		"product_name": productName,
	}, nil
}

func (r *mysqlBatchRepository) GetAvailableForProduct(productId int) ([]domain.ProductBatch, error) {
	rows, err := r.db.Query("SELECT id, product_id, batch_number, expiry_date, current_stock, purchase_price, selling_price, is_verified FROM product_batches WHERE product_id = ? AND current_stock > 0 AND is_verified = 1 ORDER BY expiry_date ASC", productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []domain.ProductBatch
	for rows.Next() {
		var b domain.ProductBatch
		if err := rows.Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, nil
}

func (r *mysqlBatchRepository) SetStock(id, stock int) error {
	_, err := r.db.Exec("UPDATE product_batches SET current_stock = ? WHERE id = ?", stock, id)
	return err
}

func (r *mysqlBatchRepository) GetExpiringCount(days int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM product_batches WHERE expiry_date <= date('now', '+' || ? || ' days') AND expiry_date >= date('now') AND current_stock > 0", days).Scan(&count)
	return count, err
}

func (r *mysqlBatchRepository) GetExpiredCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM product_batches WHERE expiry_date < date('now') AND current_stock > 0").Scan(&count)
	return count, err
}

func (r *mysqlBatchRepository) GetExpiringBatches(days int) ([]map[string]interface{}, error) {
	query := `
		SELECT b.*, p.name as product_name 
		FROM product_batches b 
		JOIN products p ON b.product_id = p.id 
		WHERE b.expiry_date <= date('now', '+' || ? || ' days') 
		AND b.expiry_date >= date('now') 
		AND b.current_stock > 0
		ORDER BY b.expiry_date ASC
	`
	rows, err := r.db.Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var b domain.ProductBatch
		var productName string
		err := rows.Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified, &b.CreatedAt, &productName)
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
	err := r.db.QueryRow("SELECT id, product_id, batch_number, expiry_date, current_stock, purchase_price, selling_price, is_verified, created_at FROM product_batches WHERE id = ?", id).
		Scan(&b.ID, &b.ProductID, &b.BatchNumber, &b.ExpiryDate, &b.CurrentStock, &b.PurchasePrice, &b.SellingPrice, &b.IsVerified, &b.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}
