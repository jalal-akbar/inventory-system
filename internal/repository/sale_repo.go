package repository

import (
	"database/sql"
	"inventory-system/internal/domain"
)

type SaleRepository interface {
	CreateSale(s *domain.Sale) (int, error)
	CreateSaleItem(i *domain.SaleItem) error
	FindByID(id int) (*domain.Sale, error)
	GetSaleItems(saleID int) ([]domain.SaleItem, error)
	SetStatus(id int, status string) error
	CreateVoidRequest(id int, reason string, requestedBy int) error
	GetPendingVoidCount() (int, error)
	WithTx(tx *sql.Tx) SaleRepository
}

type mysqlSaleRepository struct {
	db DBExecutor
}

func NewSaleRepository(db *sql.DB) SaleRepository {
	return &mysqlSaleRepository{db: db}
}

func (r *mysqlSaleRepository) getDB() DBExecutor {
	return r.db
}

func (r *mysqlSaleRepository) WithTx(tx *sql.Tx) SaleRepository {
	return &mysqlSaleRepository{db: tx}
}

func (r *mysqlSaleRepository) CreateSale(s *domain.Sale) (int, error) {
	res, err := r.getDB().Exec(`
		INSERT INTO sales (user_id, total_amount, profit, discount, payment_method, customer_name, doctor_name, prescription_number, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active', CURRENT_TIMESTAMP)`,
		s.UserID, s.TotalAmount, s.Profit, s.Discount, s.PaymentMethod, s.CustomerName, s.DoctorName, s.PrescriptionNumber)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *mysqlSaleRepository) CreateSaleItem(i *domain.SaleItem) error {
	_, err := r.getDB().Exec(`
		INSERT INTO sale_items (sale_id, product_id, batch_id, quantity, price, subtotal, sale_unit, sub_unit, items_per_unit)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		i.SaleID, i.ProductID, i.BatchID, i.Quantity, i.Price, i.Subtotal, i.SaleUnit, i.SubUnit, i.ItemsPerUnit)
	return err
}

func (r *mysqlSaleRepository) FindByID(id int) (*domain.Sale, error) {
	s := &domain.Sale{}
	var custName, docName, scriptNum, voidReason sql.NullString
	err := r.getDB().QueryRow(`
		SELECT id, user_id, total_amount, profit, discount, payment_method, customer_name, doctor_name, prescription_number, status, void_reason, void_requested_by, created_at
		FROM sales WHERE id = ?`, id).Scan(
		&s.ID, &s.UserID, &s.TotalAmount, &s.Profit, &s.Discount, &s.PaymentMethod, &custName, &docName, &scriptNum, &s.Status, &voidReason, &s.VoidRequestedBy, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if custName.Valid {
		s.CustomerName = &custName.String
	}
	if docName.Valid {
		s.DoctorName = &docName.String
	}
	if scriptNum.Valid {
		s.PrescriptionNumber = &scriptNum.String
	}
	if voidReason.Valid {
		s.VoidReason = &voidReason.String
	}

	return s, nil
}

func (r *mysqlSaleRepository) GetSaleItems(saleID int) ([]domain.SaleItem, error) {
	rows, err := r.getDB().Query("SELECT id, sale_id, product_id, batch_id, quantity, price, subtotal, sale_unit, sub_unit, items_per_unit FROM sale_items WHERE sale_id = ?", saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.SaleItem
	for rows.Next() {
		var i domain.SaleItem
		if err := rows.Scan(&i.ID, &i.SaleID, &i.ProductID, &i.BatchID, &i.Quantity, &i.Price, &i.Subtotal, &i.SaleUnit, &i.SubUnit, &i.ItemsPerUnit); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (r *mysqlSaleRepository) SetStatus(id int, status string) error {
	_, err := r.getDB().Exec("UPDATE sales SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *mysqlSaleRepository) CreateVoidRequest(id int, reason string, requestedBy int) error {
	_, err := r.getDB().Exec("UPDATE sales SET status = 'pending_void', void_reason = ?, void_requested_by = ? WHERE id = ?", reason, requestedBy, id)
	return err
}

func (r *mysqlSaleRepository) GetPendingVoidCount() (int, error) {
	var count int
	err := r.getDB().QueryRow("SELECT COUNT(*) FROM sales WHERE status = 'pending_void'").Scan(&count)
	return count, err
}
