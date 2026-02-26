package repository

import (
	"database/sql"
	"fmt"
	"inventory-system/internal/domain"
)

type ProductRepository interface {
	FindByID(id int) (*domain.Product, error)
	SearchWithStock(search, filter string) ([]map[string]interface{}, error)
	QuickSearch(query string) ([]map[string]interface{}, error)
	Create(p *domain.Product) (int, error)
	UpdateFull(id int, p *domain.Product, updatePrices bool) error
	UpdatePrices(id int, purchasePrice, sellingPrice float64) error
	Verify(id int) error
	SoftDelete(id int) error
	GetLowStockCount() (int, error)
	GetPendingCount() (int, error)
	GetPendingPricesCount() (int, error)
	GetActiveProducts() ([]domain.Product, error)
	FindWithDetails(id int) (map[string]interface{}, error)
	GetDetailsWithRelations(id int) (map[string]interface{}, error)
	GetPendingGrouped() ([]map[string]interface{}, error)
	GetRecent(limit int) ([]map[string]interface{}, error)
	GetBestSellers(limit int) ([]map[string]interface{}, error)
	SearchWithAllBatches(search, filter string) ([]map[string]interface{}, error)
	WithTx(tx *sql.Tx) ProductRepository
}

type mysqlProductRepository struct {
	db DBExecutor
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &mysqlProductRepository{db: db}
}

func (r *mysqlProductRepository) getDB() DBExecutor {
	return r.db
}

func (r *mysqlProductRepository) WithTx(tx *sql.Tx) ProductRepository {
	return &mysqlProductRepository{db: tx}
}

func (r *mysqlProductRepository) FindByID(id int) (*domain.Product, error) {
	p := &domain.Product{}
	err := r.getDB().QueryRow("SELECT id, name, sku_code, category, legal_category, therapeutic_class, unit, items_per_unit, storage_location, purchase_price, selling_price, min_stock, status, is_verified, created_at FROM products WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.SKUCode, &p.Category, &p.LegalCategory, &p.TherapeuticClass, &p.Unit, &p.ItemsPerUnit, &p.StorageLocation, &p.PurchasePrice, &p.SellingPrice, &p.MinStock, &p.Status, &p.IsVerified, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *mysqlProductRepository) SearchWithStock(search, filter string) ([]map[string]interface{}, error) {
	query := `SELECT p.id, p.name, p.sku_code, p.category, p.legal_category, p.therapeutic_class, p.unit, p.items_per_unit, p.storage_location, 
	                 p.purchase_price, p.selling_price, p.status, p.is_verified,
	                 COALESCE(SUM(b.current_stock), 0) as total_stock 
	          FROM products p 
	          LEFT JOIN product_batches b ON p.id = b.product_id AND b.is_verified = 1
	          WHERE p.status = 'active'`

	var params []interface{}
	if filter != "" && filter != "low_stock" && filter != "expiring" && filter != "expired" && filter != "pending" {
		query += " AND (p.category = ? OR p.therapeutic_class = ? OR p.legal_category = ?)"
		params = append(params, filter, filter, filter)
	}

	if search != "" {
		query += " AND (p.name LIKE ? OR p.sku_code = ? OR p.category LIKE ? OR p.legal_category LIKE ? OR p.therapeutic_class LIKE ? OR p.storage_location LIKE ?)"
		s := "%" + search + "%"
		params = append(params, s, search, s, s, s, s)
	}

	switch filter {
	case "pending":
		query += " AND p.is_verified = 0"
	case "expired":
		query += " AND EXISTS (SELECT 1 FROM product_batches pb WHERE pb.product_id = p.id AND pb.expiry_date < date('now') AND pb.current_stock > 0)"
	case "expiring":
		query += " AND EXISTS (SELECT 1 FROM product_batches pb WHERE pb.product_id = p.id AND pb.expiry_date BETWEEN date('now') AND date('now', '+90 days') AND pb.current_stock > 0)"
	}

	query += " GROUP BY p.id"

	if filter == "low_stock" {
		query += " HAVING total_stock < p.min_stock"
	}

	query += " ORDER BY p.name ASC"

	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var p domain.Product
		var totalStock int
		if err := rows.Scan(&p.ID, &p.Name, &p.SKUCode, &p.Category, &p.LegalCategory, &p.TherapeuticClass, &p.Unit, &p.ItemsPerUnit, &p.StorageLocation, &p.PurchasePrice, &p.SellingPrice, &p.Status, &p.IsVerified, &totalStock); err != nil {
			return nil, err
		}
		products = append(products, map[string]interface{}{
			"id":                p.ID,
			"name":              p.Name,
			"sku_code":          p.SKUCode,
			"category":          p.Category,
			"legal_category":    p.LegalCategory,
			"therapeutic_class": p.TherapeuticClass,
			"unit":              p.Unit,
			"items_per_unit":    p.ItemsPerUnit,
			"storage_location":  p.StorageLocation,
			"purchase_price":    p.PurchasePrice,
			"selling_price":     p.SellingPrice,
			"status":            p.Status,
			"is_verified":       p.IsVerified,
			"total_stock":       totalStock,
		})
	}
	return products, nil
}

func (r *mysqlProductRepository) QuickSearch(q string) ([]map[string]interface{}, error) {
	query := `
		SELECT p.id, p.name, p.sku_code, p.category, p.legal_category, p.therapeutic_class, p.unit, p.items_per_unit, p.selling_price,
		       COALESCE(SUM(b.current_stock), 0) as total_stock,
		       MIN(CASE WHEN b.current_stock > 0 THEN b.expiry_date ELSE NULL END) as nearest_expiry
		FROM products p
		LEFT JOIN product_batches b ON p.id = b.product_id
		WHERE p.status = 'active' AND (p.name LIKE ? OR p.sku_code = ?)
		GROUP BY p.id
		LIMIT 10
	`
	rows, err := r.db.Query(query, "%"+q+"%", q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, itemsPerUnit, totalStock int
		var name, sku, category, legalCategory, therapeuticClass, unit string
		var sellingPrice float64
		var nearestExpiry sql.NullString
		if err := rows.Scan(&id, &name, &sku, &category, &legalCategory, &therapeuticClass, &unit, &itemsPerUnit, &sellingPrice, &totalStock, &nearestExpiry); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id":                id,
			"name":              name,
			"sku_code":          sku,
			"category":          category,
			"legal_category":    legalCategory,
			"therapeutic_class": therapeuticClass,
			"unit":              unit,
			"items_per_unit":    itemsPerUnit,
			"selling_price":     sellingPrice,
			"total_stock":       totalStock,
			"nearest_expiry":    nearestExpiry.String,
		})
	}
	return results, nil
}

func (r *mysqlProductRepository) Create(p *domain.Product) (int, error) {
	res, err := r.db.Exec("INSERT INTO products (name, sku_code, category, legal_category, therapeutic_class, unit, items_per_unit, storage_location, purchase_price, selling_price, min_stock, status, is_verified, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		p.Name, p.SKUCode, p.Category, p.LegalCategory, p.TherapeuticClass, p.Unit, p.ItemsPerUnit, p.StorageLocation, p.PurchasePrice, p.SellingPrice, p.MinStock, p.Status, p.IsVerified)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *mysqlProductRepository) UpdateFull(id int, p *domain.Product, isAdmin bool) error {
	query := "UPDATE products SET name = ?, sku_code = ?, category = ?, legal_category = ?, therapeutic_class = ?, unit = ?, items_per_unit = ?, storage_location = ?, min_stock = ?, is_verified = ?"
	params := []interface{}{p.Name, p.SKUCode, p.Category, p.LegalCategory, p.TherapeuticClass, p.Unit, p.ItemsPerUnit, p.StorageLocation, p.MinStock}

	isVerified := 0
	if isAdmin {
		isVerified = 1
	}
	params = append(params, isVerified)

	if isAdmin {
		query += ", purchase_price = ?, selling_price = ?"
		params = append(params, p.PurchasePrice, p.SellingPrice)
	}

	query += " WHERE id = ?"
	params = append(params, id)

	_, err := r.db.Exec(query, params...)
	return err
}

func (r *mysqlProductRepository) UpdatePrices(id int, purchasePrice, sellingPrice float64) error {
	_, err := r.db.Exec("UPDATE products SET purchase_price = ?, selling_price = ? WHERE id = ?", purchasePrice, sellingPrice, id)
	return err
}

func (r *mysqlProductRepository) Verify(id int) error {
	_, err := r.db.Exec("UPDATE products SET is_verified = 1 WHERE id = ?", id)
	return err
}

func (r *mysqlProductRepository) SoftDelete(id int) error {
	_, err := r.db.Exec("UPDATE products SET status = 'inactive' WHERE id = ?", id)
	return err
}

func (r *mysqlProductRepository) GetLowStockCount() (int, error) {
	query := `SELECT COUNT(*) FROM (
		SELECT p.id
		FROM products p 
		LEFT JOIN product_batches pb ON p.id = pb.product_id 
		WHERE p.status = 'active'
		GROUP BY p.id 
		HAVING SUM(IFNULL(pb.current_stock, 0)) < p.min_stock
	) as low_stock_products`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

func (r *mysqlProductRepository) GetPendingCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM products WHERE is_verified = 0 AND status = 'active'").Scan(&count)
	return count, err
}

func (r *mysqlProductRepository) GetPendingPricesCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM products WHERE selling_price = 0 AND status = 'active'").Scan(&count)
	return count, err
}

func (r *mysqlProductRepository) GetActiveProducts() ([]domain.Product, error) {
	rows, err := r.db.Query("SELECT id, name, sku_code, category, legal_category, therapeutic_class, unit, items_per_unit, storage_location, purchase_price, selling_price, min_stock, status, is_verified FROM products WHERE status = 'active' ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKUCode, &p.Category, &p.LegalCategory, &p.TherapeuticClass, &p.Unit, &p.ItemsPerUnit, &p.StorageLocation, &p.PurchasePrice, &p.SellingPrice, &p.MinStock, &p.Status, &p.IsVerified); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *mysqlProductRepository) FindWithDetails(id int) (map[string]interface{}, error) {
	row := r.db.QueryRow("SELECT id, name, sku_code, storage_location, selling_price FROM products WHERE id = ?", id)
	var pID int
	var name, sku, location string
	var price float64
	if err := row.Scan(&pID, &name, &sku, &location, &price); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":               pID,
		"name":             name,
		"sku_code":         sku,
		"storage_location": location,
		"selling_price":    price,
	}, nil
}

func (r *mysqlProductRepository) GetDetailsWithRelations(id int) (map[string]interface{}, error) {
	query := `
		SELECT p.id, p.name, p.sku_code, p.category, p.legal_category, p.therapeutic_class, p.unit, p.items_per_unit, p.storage_location,
		       p.purchase_price, p.selling_price, p.min_stock, p.is_verified,
		       b.batch_number, b.expiry_date, b.current_stock as initial_stock_pcs,
		       u.username as staff_name,
		       se.created_at as request_date
		FROM products p
		LEFT JOIN product_batches b ON p.id = b.product_id
		LEFT JOIN stock_entries se ON p.id = se.product_id AND b.id = se.batch_id
		LEFT JOIN users u ON se.requested_by = u.id
		WHERE p.id = ?
		LIMIT 1
	`
	row := r.db.QueryRow(query, id)
	var p domain.Product
	var batchNumber, expiryDate, staffName sql.NullString
	var initialStock int
	var requestDate sql.NullTime

	err := row.Scan(&p.ID, &p.Name, &p.SKUCode, &p.Category, &p.LegalCategory, &p.TherapeuticClass, &p.Unit, &p.ItemsPerUnit, &p.StorageLocation,
		&p.PurchasePrice, &p.SellingPrice, &p.MinStock, &p.IsVerified,
		&batchNumber, &expiryDate, &initialStock, &staffName, &requestDate)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"product":       p,
		"batch_number":  batchNumber.String,
		"expiry_date":   expiryDate.String,
		"initial_stock": initialStock,
		"staff_name":    staffName.String,
		"request_date":  requestDate.Time,
	}, nil
}
func (r *mysqlProductRepository) GetPendingGrouped() ([]map[string]interface{}, error) {
	// Reconstructing the logic: UNION between products/batches/stock and void requests
	query := `
		(SELECT 
			p.id as product_id, p.name, p.category, p.legal_category, p.therapeutic_class, p.unit, p.purchase_price, p.selling_price,
			'product_group' as type,
			"" as void_reason,
			GROUP_CONCAT(pb.id || ':' || pb.batch_number || ':' || pb.expiry_date || ':' || pb.current_stock, '||') as batches,
			GROUP_CONCAT(se.id, '||') as stock_entries
		FROM products p
		LEFT JOIN product_batches pb ON p.id = pb.product_id AND pb.is_verified = 0
		LEFT JOIN stock_entries se ON p.id = se.product_id AND se.is_verified = 0
		WHERE p.is_verified = 0 OR pb.id IS NOT NULL OR se.id IS NOT NULL
		GROUP BY p.id)
		UNION ALL
		(SELECT 
			s.id as product_id, "" as name, "" as category, "" as legal_category, "" as therapeutic_class, "" as unit, 0 as purchase_price, s.total_amount as selling_price,
			'void_request' as type,
			s.void_reason,
			"" as batches,
			"" as stock_entries
		FROM sales s
		WHERE s.status = 'pending_void')
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var productID int
		var name, category, legalCategory, therapeuticClass, unit, voidReason, batches, stockEntries, itemType string
		var purchasePrice, sellingPrice float64

		if err := rows.Scan(&productID, &name, &category, &legalCategory, &therapeuticClass, &unit, &purchasePrice, &sellingPrice, &itemType, &voidReason, &batches, &stockEntries); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"product_id":        productID,
			"name":              name,
			"category":          category,
			"legal_category":    legalCategory,
			"therapeutic_class": therapeuticClass,
			"unit":              unit,
			"purchase_price":    purchasePrice,
			"selling_price":     sellingPrice,
			"type":              itemType,
			"void_reason":       voidReason,
			"batches":           batches,
			"stock_entries":     stockEntries,
		})
	}
	return results, nil
}

func (r *mysqlProductRepository) GetRecent(limit int) ([]map[string]interface{}, error) {
	query := `SELECT id, name, sku_code, category, legal_category, therapeutic_class, selling_price,
	                 (SELECT COALESCE(SUM(current_stock), 0) FROM product_batches WHERE product_id = p.id) as total_stock
	          FROM products p
	          WHERE status = 'active'
	          ORDER BY created_at DESC
	          LIMIT ?`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, totalStock int
		var name, sku, category, legalCategory, therapeuticClass string
		var price float64
		if err := rows.Scan(&id, &name, &sku, &category, &legalCategory, &therapeuticClass, &price, &totalStock); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id":                id,
			"name":              name,
			"sku_code":          sku,
			"category":          category,
			"legal_category":    legalCategory,
			"therapeutic_class": therapeuticClass,
			"selling_price":     price,
			"total_stock":       totalStock,
		})
	}
	return results, nil
}

func (r *mysqlProductRepository) GetBestSellers(limit int) ([]map[string]interface{}, error) {
	// First, try to get products by sales quantity
	query := `SELECT p.id, p.name, p.sku_code, p.category, p.legal_category, p.therapeutic_class, p.unit, p.items_per_unit, p.storage_location, 
	                 p.purchase_price, p.selling_price, p.status, p.is_verified,
	                 COALESCE(SUM(si.quantity), 0) as total_sold,
	                 (SELECT COALESCE(SUM(current_stock), 0) FROM product_batches WHERE product_id = p.id AND is_verified = 1) as total_stock
	          FROM products p
	          JOIN sale_items si ON p.id = si.product_id
	          JOIN sales s ON si.sale_id = s.id
	          WHERE p.status = 'active' AND p.is_verified = 1 AND s.status = 'active'
	          GROUP BY p.id
	          ORDER BY total_sold DESC
	          LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var p domain.Product
		var totalSold, totalStock int
		if err := rows.Scan(&p.ID, &p.Name, &p.SKUCode, &p.Category, &p.LegalCategory, &p.TherapeuticClass, &p.Unit, &p.ItemsPerUnit, &p.StorageLocation, &p.PurchasePrice, &p.SellingPrice, &p.Status, &p.IsVerified, &totalSold, &totalStock); err != nil {
			return nil, err
		}
		products = append(products, map[string]interface{}{
			"id":                p.ID,
			"name":              p.Name,
			"sku_code":          p.SKUCode,
			"category":          p.Category,
			"legal_category":    p.LegalCategory,
			"therapeutic_class": p.TherapeuticClass,
			"unit":              p.Unit,
			"items_per_unit":    p.ItemsPerUnit,
			"storage_location":  p.StorageLocation,
			"purchase_price":    p.PurchasePrice,
			"selling_price":     p.SellingPrice,
			"status":            p.Status,
			"is_verified":       p.IsVerified,
			"total_stock":       totalStock,
			"total_sold":        totalSold,
		})
	}

	// If we have enough best sellers, return them
	if len(products) >= limit {
		return products, nil
	}

	// Otherwise, pad with recently added products that are not yet in the best sellers list
	alreadyIn := make(map[int]bool)
	for _, p := range products {
		alreadyIn[p["id"].(int)] = true
	}

	remainingLimit := limit - len(products)
	recentQuery := `SELECT p.id, p.name, p.sku_code, p.category, p.legal_category, p.therapeutic_class, p.unit, p.items_per_unit, p.storage_location, 
	                       p.purchase_price, p.selling_price, p.status, p.is_verified,
	                       (SELECT COALESCE(SUM(current_stock), 0) FROM product_batches WHERE product_id = p.id AND is_verified = 1) as total_stock
	                FROM products p
	                WHERE p.status = 'active' AND p.is_verified = 1`

	if len(alreadyIn) > 0 {
		ids := ""
		i := 0
		for id := range alreadyIn {
			if i > 0 {
				ids += ","
			}
			ids += fmt.Sprintf("%d", id)
			i++
		}
		recentQuery += fmt.Sprintf(" AND p.id NOT IN (%s)", ids)
	}

	recentQuery += " ORDER BY p.created_at DESC LIMIT ?"

	recentRows, err := r.db.Query(recentQuery, remainingLimit)
	if err != nil {
		// If padding fails, just return what we have
		return products, nil
	}
	defer recentRows.Close()

	for recentRows.Next() {
		var p domain.Product
		var totalStock int
		if err := recentRows.Scan(&p.ID, &p.Name, &p.SKUCode, &p.Category, &p.LegalCategory, &p.TherapeuticClass, &p.Unit, &p.ItemsPerUnit, &p.StorageLocation, &p.PurchasePrice, &p.SellingPrice, &p.Status, &p.IsVerified, &totalStock); err != nil {
			return products, nil
		}
		products = append(products, map[string]interface{}{
			"id":                p.ID,
			"name":              p.Name,
			"sku_code":          p.SKUCode,
			"category":          p.Category,
			"legal_category":    p.LegalCategory,
			"therapeutic_class": p.TherapeuticClass,
			"unit":              p.Unit,
			"items_per_unit":    p.ItemsPerUnit,
			"storage_location":  p.StorageLocation,
			"purchase_price":    p.PurchasePrice,
			"selling_price":     p.SellingPrice,
			"status":            p.Status,
			"is_verified":       p.IsVerified,
			"total_stock":       totalStock,
			"total_sold":        0,
		})
	}

	return products, nil
}

func (r *mysqlProductRepository) SearchWithAllBatches(search, filter string) ([]map[string]interface{}, error) {
	query := `SELECT p.id, p.name, p.sku_code, p.category, p.legal_category, p.therapeutic_class, p.unit, p.items_per_unit,
	                 b.id as batch_id, b.batch_number, b.expiry_date, b.current_stock
	          FROM products p 
	          JOIN product_batches b ON p.id = b.product_id
	          WHERE p.status = 'active' AND b.is_verified = 1`

	var params []interface{}
	if search != "" {
		query += " AND (p.name LIKE ? OR p.sku_code = ? OR p.category LIKE ?)"
		s := "%" + search + "%"
		params = append(params, s, search, s)
	}

	query += " ORDER BY p.name ASC, b.expiry_date ASC"

	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, batchID, currentStock, itemsPerUnit int
		var name, sku, category, legalCategory, therapeuticClass, unit, batchNumber, expiryDate string
		if err := rows.Scan(&id, &name, &sku, &category, &legalCategory, &therapeuticClass, &unit, &itemsPerUnit, &batchID, &batchNumber, &expiryDate, &currentStock); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id":                id,
			"name":              name,
			"sku_code":          sku,
			"category":          category,
			"legal_category":    legalCategory,
			"therapeutic_class": therapeuticClass,
			"unit":              unit,
			"items_per_unit":    itemsPerUnit,
			"batch_id":          batchID,
			"batch_number":      batchNumber,
			"expiry_date":       expiryDate,
			"current_stock":     currentStock,
		})
	}
	return results, nil
}
