package repository

import (
	"database/sql"
	"inventory-system/internal/domain"
)

type StockEntryRepository interface {
	VerifyByProduct(productId int) error
	VerifyByBatch(batchId int) error
	Verify(id int) error
	Reject(id int) error
	GetPendingCount() (int, error)
	Create(se *domain.StockEntry) error
	WithTx(tx *sql.Tx) StockEntryRepository
}

type mysqlStockEntryRepository struct {
	db DBExecutor
}

func NewStockEntryRepository(db *sql.DB) StockEntryRepository {
	return &mysqlStockEntryRepository{db: db}
}

func (r *mysqlStockEntryRepository) getDB() DBExecutor {
	return r.db
}

func (r *mysqlStockEntryRepository) WithTx(tx *sql.Tx) StockEntryRepository {
	return &mysqlStockEntryRepository{db: tx}
}

func (r *mysqlStockEntryRepository) VerifyByProduct(productId int) error {
	_, err := r.getDB().Exec("UPDATE stock_entries SET is_verified = 1 WHERE product_id = ? AND is_verified = 0", productId)
	return err
}

func (r *mysqlStockEntryRepository) VerifyByBatch(batchId int) error {
	_, err := r.getDB().Exec("UPDATE stock_entries SET is_verified = 1 WHERE batch_id = ? AND is_verified = 0", batchId)
	return err
}

func (r *mysqlStockEntryRepository) Verify(id int) error {
	_, err := r.getDB().Exec("UPDATE stock_entries SET is_verified = 1 WHERE id = ?", id)
	return err
}

func (r *mysqlStockEntryRepository) Reject(id int) error {
	_, err := r.getDB().Exec("UPDATE stock_entries SET status = 'rejected' WHERE id = ?", id)
	return err
}

func (r *mysqlStockEntryRepository) GetPendingCount() (int, error) {
	var count int
	err := r.getDB().QueryRow("SELECT COUNT(*) FROM stock_entries WHERE status = 'approved' AND is_verified = 0").Scan(&count)
	return count, err
}

func (r *mysqlStockEntryRepository) Create(se *domain.StockEntry) error {
	_, err := r.getDB().Exec("INSERT INTO stock_entries (product_id, batch_id, quantity, status, is_verified, requested_by, created_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		se.ProductID, se.BatchID, se.Quantity, se.Status, se.IsVerified, se.RequestedBy)
	return err
}
