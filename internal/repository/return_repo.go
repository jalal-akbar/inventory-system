package repository

import (
	"database/sql"
	"inventory-system/internal/domain"
)

type ReturnRepository interface {
	CreateReturn(r *domain.Return) (int, error)
	CreateReturnItem(i *domain.ReturnItem) error
	GetReturnsBySaleID(saleID int) ([]domain.Return, error)
	GetReturnedQtyBySaleItemID(saleItemID int) (int, error)
	GetReturnWithDetails(returnID int) (*domain.ReturnDetail, error)
}

type mysqlReturnRepository struct {
	db *sql.DB
}

func NewReturnRepository(db *sql.DB) ReturnRepository {
	return &mysqlReturnRepository{db: db}
}

func (r *mysqlReturnRepository) CreateReturn(ret *domain.Return) (int, error) {
	res, err := r.db.Exec(`
		INSERT INTO returns (sale_id, user_id, total_refund, reason, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		ret.SaleID, ret.UserID, ret.TotalRefund, ret.Reason)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *mysqlReturnRepository) CreateReturnItem(i *domain.ReturnItem) error {
	_, err := r.db.Exec(`
		INSERT INTO return_items (return_id, sale_item_id, quantity, refund_amount, condition_status)
		VALUES (?, ?, ?, ?, ?)`,
		i.ReturnID, i.SaleItemID, i.Quantity, i.RefundAmount, i.ConditionStatus)
	return err
}

func (r *mysqlReturnRepository) GetReturnsBySaleID(saleID int) ([]domain.Return, error) {
	rows, err := r.db.Query(`
		SELECT id, sale_id, user_id, total_refund, reason, created_at
		FROM returns WHERE sale_id = ?`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var returns []domain.Return
	for rows.Next() {
		var ret domain.Return
		var reason sql.NullString
		if err := rows.Scan(&ret.ID, &ret.SaleID, &ret.UserID, &ret.TotalRefund, &reason, &ret.CreatedAt); err != nil {
			return nil, err
		}
		if reason.Valid {
			ret.Reason = &reason.String
		}
		returns = append(returns, ret)
	}
	return returns, nil
}

func (r *mysqlReturnRepository) GetReturnedQtyBySaleItemID(saleItemID int) (int, error) {
	var total int
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(quantity), 0)
		FROM return_items WHERE sale_item_id = ?`, saleItemID).Scan(&total)
	return total, err
}

func (r *mysqlReturnRepository) GetReturnWithDetails(returnID int) (*domain.ReturnDetail, error) {
	detail := &domain.ReturnDetail{}
	var reason sql.NullString

	err := r.db.QueryRow(`
		SELECT id, sale_id, user_id, total_refund, reason, created_at
		FROM returns WHERE id = ?`, returnID).Scan(
		&detail.ID, &detail.SaleID, &detail.UserID, &detail.TotalRefund, &reason, &detail.CreatedAt)
	if err != nil {
		return nil, err
	}
	if reason.Valid {
		detail.Reason = &reason.String
	}

	rows, err := r.db.Query(`
		SELECT ri.id, ri.return_id, ri.sale_item_id, ri.quantity, ri.refund_amount, ri.condition_status,
		       p.name, pb.batch_number
		FROM return_items ri
		JOIN sale_items si ON ri.sale_item_id = si.id
		JOIN products p ON si.product_id = p.id
		JOIN product_batches pb ON si.batch_id = pb.id
		WHERE ri.return_id = ?`, returnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.ReturnItemDetail
		if err := rows.Scan(&item.ID, &item.ReturnID, &item.SaleItemID, &item.Quantity, &item.RefundAmount, &item.ConditionStatus, &item.ProductName, &item.BatchNumber); err != nil {
			return nil, err
		}
		detail.Items = append(detail.Items, item)
	}

	return detail, nil
}
