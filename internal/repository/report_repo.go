package repository

import (
	"database/sql"
	"time"
)

type FinancialSummary struct {
	Count   int     `json:"count"`
	Revenue float64 `json:"revenue"`
	Profit  float64 `json:"profit"`
}

type PaymentBreakdown struct {
	PaymentMethod string  `json:"payment_method"`
	Total         float64 `json:"total"`
}

type StaffBreakdown struct {
	Username     string  `json:"username"`
	Transactions int     `json:"transactions"`
	Revenue      float64 `json:"revenue"`
}

type PsychotropicReportRow struct {
	SaleDate           time.Time `json:"sale_date"`
	ProductName        string    `json:"product_name"`
	BatchNumber        string    `json:"batch_number"`
	Quantity           int       `json:"quantity"`
	CustomerName       string    `json:"customer_name"`
	DoctorName         string    `json:"doctor_name"`
	PrescriptionNumber string    `json:"prescription_number"`
}

type StockMutationRow struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Unit         string `json:"unit"`
	ItemsPerUnit int    `json:"items_per_unit"`
	StokAwal     int    `json:"stok_awal"`
	Masuk        int    `json:"masuk"`
	Keluar       int    `json:"keluar"`
}

type TodaySummary struct {
	Transactions      int     `json:"transactions"`
	Revenue           float64 `json:"revenue"`
	PsychotropicSales int     `json:"psychotropic_sales"`
}

type ReportRepository interface {
	GetFinancialSummary(startDate, endDate, category string) (*FinancialSummary, error)
	GetPaymentMethodBreakdown(startDate, endDate, category string) ([]PaymentBreakdown, error)
	GetStaffBreakdown(startDate, endDate, category string) ([]StaffBreakdown, error)
	GetRecentSales(startDate, endDate, category string, limit int) ([]map[string]interface{}, error)
	GetHistory(startDate, endDate string) ([]map[string]interface{}, error)
	GetPsychotropicReport(startDate, endDate string) ([]PsychotropicReportRow, error)
	GetStockMutation(startDate, endDate string) ([]StockMutationRow, error)
	CountExpiringSoon(days int) (int, error)
	GetTodaySummary(today string) (*TodaySummary, error)
	GetYesterdaySummary(yesterday string) (*TodaySummary, error)
	GetProfitSummary(startDate, endDate string) (float64, error)
	GetWeeklySales(dates []string) ([]string, []float64, error)
	GetExpiringCount(days int, now string) (int, error)
	GetExpiredCount(now string) (int, error)
}

type mysqlReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) ReportRepository {
	return &mysqlReportRepository{db: db}
}

func (r *mysqlReportRepository) GetFinancialSummary(startDate, endDate, category string) (*FinancialSummary, error) {
	whereClause := "WHERE DATE(created_at) BETWEEN ? AND ? AND status = 'active'"
	params := []interface{}{startDate, endDate}

	if category != "" {
		whereClause += " AND EXISTS (SELECT 1 FROM sale_items si JOIN products p ON si.product_id = p.id WHERE si.sale_id = sales.id AND p.category = ?)"
		params = append(params, category)
	}

	query := "SELECT COUNT(*) as count, COALESCE(SUM(total_amount), 0) as revenue, COALESCE(SUM(profit), 0) as profit FROM sales " + whereClause
	var summary FinancialSummary
	err := r.db.QueryRow(query, params...).Scan(&summary.Count, &summary.Revenue, &summary.Profit)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *mysqlReportRepository) GetPaymentMethodBreakdown(startDate, endDate, category string) ([]PaymentBreakdown, error) {
	whereClause := "WHERE DATE(created_at) BETWEEN ? AND ? AND status = 'active'"
	params := []interface{}{startDate, endDate}

	if category != "" {
		whereClause += " AND EXISTS (SELECT 1 FROM sale_items si JOIN products p ON si.product_id = p.id WHERE si.sale_id = sales.id AND p.category = ?)"
		params = append(params, category)
	}

	query := "SELECT payment_method, COALESCE(SUM(total_amount), 0) as total FROM sales " + whereClause + " GROUP BY payment_method"
	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PaymentBreakdown
	for rows.Next() {
		var pb PaymentBreakdown
		if err := rows.Scan(&pb.PaymentMethod, &pb.Total); err != nil {
			return nil, err
		}
		result = append(result, pb)
	}
	return result, nil
}

func (r *mysqlReportRepository) GetStaffBreakdown(startDate, endDate, category string) ([]StaffBreakdown, error) {
	whereClause := "WHERE date(s.created_at) BETWEEN ? AND ? AND s.status = 'active'"
	params := []interface{}{startDate, endDate}

	if category != "" {
		whereClause += " AND EXISTS (SELECT 1 FROM sale_items si JOIN products p ON si.product_id = p.id WHERE si.sale_id = s.id AND p.category = ?)"
		params = append(params, category)
	}

	query := "SELECT u.username, COUNT(s.id) as transactions, COALESCE(SUM(s.total_amount), 0) as revenue " +
		"FROM sales s " +
		"JOIN users u ON s.user_id = u.id " +
		whereClause + " GROUP BY s.user_id"

	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []StaffBreakdown
	for rows.Next() {
		var sb StaffBreakdown
		if err := rows.Scan(&sb.Username, &sb.Transactions, &sb.Revenue); err != nil {
			return nil, err
		}
		result = append(result, sb)
	}
	return result, nil
}

func (r *mysqlReportRepository) GetRecentSales(startDate, endDate, category string, limit int) ([]map[string]interface{}, error) {
	whereClause := "WHERE date(s.created_at) BETWEEN ? AND ?"
	params := []interface{}{startDate, endDate}

	if category != "" {
		whereClause += " AND s.status = 'active' AND EXISTS (SELECT 1 FROM sale_items si JOIN products p ON si.product_id = p.id WHERE si.sale_id = s.id AND p.category = ?)"
		params = append(params, category)
	}

	query := `SELECT s.*, u.username, 
              (SELECT COUNT(*) FROM sale_items si 
               JOIN products p ON si.product_id = p.id 
               WHERE si.sale_id = s.id AND p.is_verified = 0) as unverified_count
              FROM sales s 
              JOIN users u ON s.user_id = u.id ` +
		whereClause + ` ORDER BY s.created_at DESC LIMIT ?`
	params = append(params, limit)

	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanToMaps(rows)
}

func (r *mysqlReportRepository) GetHistory(startDate, endDate string) ([]map[string]interface{}, error) {
	query := `SELECT s.*, u.username as staff_name 
              FROM sales s 
              JOIN users u ON s.user_id = u.id 
              WHERE date(s.created_at) BETWEEN ? AND ?
              ORDER BY s.created_at DESC`
	params := []interface{}{startDate, endDate}

	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanToMaps(rows)
}

func (r *mysqlReportRepository) GetPsychotropicReport(startDate, endDate string) ([]PsychotropicReportRow, error) {
	query := `SELECT 
                s.created_at as sale_date,
                p.name as product_name,
                pb.batch_number,
                si.quantity,
                s.customer_name,
                s.doctor_name,
                s.prescription_number
              FROM sales s
              JOIN sale_items si ON s.id = si.sale_id
              JOIN products p ON si.product_id = p.id
              JOIN product_batches pb ON si.batch_id = pb.id
              WHERE p.category = 'Psikotropika'
              AND date(s.created_at) BETWEEN ? AND ?
              AND s.status = 'active'
              ORDER BY s.created_at DESC`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PsychotropicReportRow
	for rows.Next() {
		var row PsychotropicReportRow
		var createdAtStr string
		if err := rows.Scan(&createdAtStr, &row.ProductName, &row.BatchNumber, &row.Quantity, &row.CustomerName, &row.DoctorName, &row.PrescriptionNumber); err != nil {
			return nil, err
		}
		row.SaleDate, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		result = append(result, row)
	}
	return result, nil
}

func (r *mysqlReportRepository) GetStockMutation(startDate, endDate string) ([]StockMutationRow, error) {
	query := `SELECT 
                p.id, 
                p.name, 
                p.unit,
                p.items_per_unit,
                (
                    COALESCE((SELECT SUM(se.quantity) FROM stock_entries se WHERE se.product_id = p.id AND se.status = 'approved' AND se.created_at < ?), 0) -
                    COALESCE((SELECT SUM(si.quantity) FROM sale_items si JOIN sales s ON si.sale_id = s.id WHERE si.product_id = p.id AND s.status = 'active' AND s.created_at < ?), 0)
                ) as stok_awal,
                COALESCE((SELECT SUM(se.quantity) FROM stock_entries se WHERE se.product_id = p.id AND se.status = 'approved' AND date(se.created_at) BETWEEN ? AND ?), 0) as masuk,
                COALESCE((SELECT SUM(si.quantity) FROM sale_items si JOIN sales s ON si.sale_id = s.id WHERE si.product_id = p.id AND s.status = 'active' AND date(s.created_at) BETWEEN ? AND ?), 0) as keluar
            FROM products p
            WHERE p.category = 'Psikotropika'
            ORDER BY p.name ASC`

	rows, err := r.db.Query(query, startDate, startDate, startDate, endDate, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []StockMutationRow
	for rows.Next() {
		var row StockMutationRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Unit, &row.ItemsPerUnit, &row.StokAwal, &row.Masuk, &row.Keluar); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}

func (r *mysqlReportRepository) CountExpiringSoon(days int) (int, error) {
	now := time.Now().Format("2006-01-02")
	query := "SELECT COUNT(*) FROM product_batches WHERE current_stock > 0 AND expiry_date <= date(?, '+' || ? || ' days') AND expiry_date >= ?"
	var count int
	err := r.db.QueryRow(query, now, days, now).Scan(&count)
	return count, err
}

func (r *mysqlReportRepository) GetTodaySummary(today string) (*TodaySummary, error) {
	var summary TodaySummary

	// Revenue and transactions
	query := "SELECT COUNT(*), COALESCE(SUM(total_amount), 0) FROM sales WHERE date(created_at) = ? AND status = 'active'"
	err := r.db.QueryRow(query, today).Scan(&summary.Transactions, &summary.Revenue)
	if err != nil {
		return nil, err
	}

	// Psychotropic sales
	query = "SELECT COUNT(*) FROM sales s JOIN sale_items si ON s.id = si.sale_id JOIN products p ON si.product_id = p.id WHERE date(s.created_at) = ? AND p.category = 'Psikotropika' AND s.status = 'active'"
	err = r.db.QueryRow(query, today).Scan(&summary.PsychotropicSales)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *mysqlReportRepository) GetWeeklySales(dates []string) ([]string, []float64, error) {
	if len(dates) == 0 {
		return nil, nil, nil
	}
	startDate := dates[0]
	endDate := dates[len(dates)-1]

	query := `
		SELECT date(created_at) as date, SUM(total_amount) as revenue 
		FROM sales 
		WHERE date(created_at) BETWEEN ? AND ? AND status = 'active'
		GROUP BY date(created_at) 
		ORDER BY date(created_at) ASC`
	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	// Map to store results
	results := make(map[string]float64)
	for rows.Next() {
		var date string
		var revenue float64
		if err := rows.Scan(&date, &revenue); err != nil {
			return nil, nil, err
		}
		if len(date) > 10 {
			date = date[:10]
		}
		results[date] = revenue
	}

	var labels []string
	var values []float64
	for _, d := range dates {
		labels = append(labels, d)
		values = append(values, results[d])
	}

	return labels, values, nil
}

func (r *mysqlReportRepository) GetYesterdaySummary(yesterday string) (*TodaySummary, error) {
	var summary TodaySummary

	query := "SELECT COUNT(*), COALESCE(SUM(total_amount), 0) FROM sales WHERE date(created_at) = ? AND status = 'active'"
	err := r.db.QueryRow(query, yesterday).Scan(&summary.Transactions, &summary.Revenue)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *mysqlReportRepository) GetProfitSummary(startDate, endDate string) (float64, error) {
	query := "SELECT COALESCE(SUM(profit), 0) FROM sales WHERE date(created_at) BETWEEN ? AND ? AND status = 'active'"
	var profit float64
	err := r.db.QueryRow(query, startDate, endDate).Scan(&profit)
	return profit, err
}

func (r *mysqlReportRepository) GetExpiringCount(days int, now string) (int, error) {
	query := "SELECT COUNT(*) FROM product_batches WHERE current_stock > 0 AND expiry_date > ? AND expiry_date <= date(?, '+' || ? || ' days')"
	var count int
	err := r.db.QueryRow(query, now, now, days).Scan(&count)
	return count, err
}

func (r *mysqlReportRepository) GetExpiredCount(now string) (int, error) {
	query := "SELECT COUNT(*) FROM product_batches WHERE current_stock > 0 AND expiry_date <= ?"
	var count int
	err := r.db.QueryRow(query, now).Scan(&count)
	return count, err
}

// Helper to scan sql.Rows into a slice of maps
func (r *mysqlReportRepository) scanToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		result = append(result, entry)
	}
	return result, nil
}
