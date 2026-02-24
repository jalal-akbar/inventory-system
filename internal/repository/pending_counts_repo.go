package repository

import (
	"database/sql"
)

type PendingCounts struct {
	Total      int
	Voids      int
	Stock      int
	Unverified int
}

type PendingCountsRepository interface {
	GetCounts() (*PendingCounts, error)
}

type mysqlPendingCountsRepository struct {
	db *sql.DB
}

func NewPendingCountsRepository(db *sql.DB) PendingCountsRepository {
	return &mysqlPendingCountsRepository{db: db}
}

func (r *mysqlPendingCountsRepository) GetCounts() (*PendingCounts, error) {
	counts := &PendingCounts{}

	// Voids
	err := r.db.QueryRow("SELECT COUNT(*) FROM sales WHERE status = 'pending_void'").Scan(&counts.Voids)
	if err != nil {
		return nil, err
	}

	// Stock (Pending Add)
	err = r.db.QueryRow("SELECT COUNT(*) FROM stock_entries WHERE status = 'approved' AND is_verified = 0").Scan(&counts.Stock)
	if err != nil {
		return nil, err
	}

	// Unverified (Products + Batches)
	var productCount, batchCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM products WHERE is_verified = 0 AND status = 'active'").Scan(&productCount)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRow("SELECT COUNT(*) FROM product_batches b JOIN products p ON b.product_id = p.id WHERE b.is_verified = 0 AND p.status = 'active'").Scan(&batchCount)
	if err != nil {
		return nil, err
	}

	counts.Unverified = productCount + batchCount
	counts.Total = counts.Voids + counts.Stock + counts.Unverified

	return counts, nil
}
