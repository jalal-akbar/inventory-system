package service

import (
	"database/sql"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"testing"

	_ "modernc.org/sqlite"
)

func TestUnitConversionLogic(t *testing.T) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create in-memory DB: %v", err)
	}
	defer db.Close()

	schema := `
	CREATE TABLE products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		sku_code TEXT UNIQUE,
		category TEXT,
		unit TEXT,
		base_unit TEXT,
		items_per_unit INTEGER,
		storage_location TEXT,
		purchase_price REAL,
		selling_price REAL,
		min_stock INTEGER,
		status TEXT,
		therapeutic_class TEXT,
		is_verified INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE product_batches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER,
		batch_number TEXT,
		expiry_date TEXT,
		initial_qty INTEGER,
		current_stock INTEGER,
		purchase_price REAL,
		selling_price REAL,
		is_verified INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE sales (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		total_amount REAL,
		profit REAL,
		discount REAL,
		payment_method TEXT,
		customer_name TEXT,
		doctor_name TEXT,
		prescription_number TEXT,
		status TEXT,
		void_reason TEXT,
		void_requested_by INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE sale_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sale_id INTEGER,
		product_id INTEGER,
		batch_id INTEGER,
		quantity INTEGER,
		price REAL,
		subtotal REAL,
		sale_unit TEXT,
		base_unit TEXT,
		items_per_unit INTEGER
	);
	CREATE TABLE activity_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		action TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE stock_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER,
		batch_id INTEGER,
		quantity INTEGER,
		status TEXT,
		is_verified INTEGER,
		requested_by INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, _ = db.Exec(schema)

	productRepo := repository.NewProductRepository(db)
	batchRepo := repository.NewBatchRepository(db)
	logRepo := repository.NewActivityLogRepository(db)
	saleRepo := repository.NewSaleRepository(db)
	entryRepo := repository.NewStockEntryRepository(db)

	saleService := NewSaleService(db, saleRepo, productRepo, batchRepo, logRepo)
	batchService := NewBatchService(db, batchRepo, entryRepo, logRepo)

	// Setup: 1 Strip = 10 Pcs. Price = 5000 / Pcs.
	pID, _ := productRepo.Create(&domain.Product{
		Name:         "Test Product",
		Unit:         "Strip",
		ItemsPerUnit: 10,
		SellingPrice: 5000,
		Status:       "active",
		IsVerified:   true,
	})

	bID, _ := batchRepo.Create(&domain.ProductBatch{
		ProductID:    pID,
		BatchNumber:  "B1",
		InitialQty:   2,    // 2 Strips
		CurrentStock: 20,   // 20 Pcs
		SellingPrice: 5000, // Price per Pcs
		IsVerified:   true,
	})

	t.Run("Sale Calculation - 1 Strip", func(t *testing.T) {
		// Buying 1 Strip = 10 Pcs.
		saleItems := []domain.SaleItem{
			{ProductID: pID, Quantity: 10},
		}
		saleID, err := saleService.ProcessSale(1, saleItems, "Cash", "Customer", "", "", 0)
		if err != nil {
			t.Fatalf("ProcessSale failed: %v", err)
		}

		sale, _ := saleRepo.FindByID(saleID)
		// 10 Pcs * 5000 = 50,000
		if sale.TotalAmount != 50000 {
			t.Errorf("Expected Total 50000, Got %.2f", sale.TotalAmount)
		}

		// Remaining stock should be 10 Pcs
		var stock int
		db.QueryRow("SELECT current_stock FROM product_batches WHERE id = ?", bID).Scan(&stock)
		if stock != 10 {
			t.Errorf("Expected Stock 10, Got %d", stock)
		}
	})

	t.Run("Inventory Check Adjustment", func(t *testing.T) {
		// Adjust back to 2 Strips (20 Pcs)
		err := batchService.PerformInventoryCheck(bID, 2, "Strip", 10, "Reset", 1)
		if err != nil {
			t.Fatalf("PerformInventoryCheck failed: %v", err)
		}

		var stock int
		db.QueryRow("SELECT current_stock FROM product_batches WHERE id = ?", bID).Scan(&stock)
		if stock != 20 {
			t.Errorf("Expected Stock 20, Got %d", stock)
		}
	})
}
