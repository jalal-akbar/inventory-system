package service

import (
	"database/sql"
	"fmt"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSaleService_ProcessAndVoid(t *testing.T) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create in-memory DB: %v", err)
	}
	defer db.Close()

	// Initialize schema for test
	schema := `
	CREATE TABLE products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		sku_code TEXT UNIQUE,
		category TEXT,
		unit TEXT,
		sub_unit TEXT,
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
		sub_unit TEXT,
		items_per_unit INTEGER
	);
	CREATE TABLE activity_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		action TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	defer db.Close()

	// Setup Repositories
	productRepo := repository.NewProductRepository(db)
	batchRepo := repository.NewBatchRepository(db)
	logRepo := repository.NewActivityLogRepository(db)
	saleRepo := repository.NewSaleRepository(db)

	service := NewSaleService(db, saleRepo, productRepo, batchRepo, logRepo)

	// 1. Setup Test Data
	sku := fmt.Sprintf("TEST-INT-POS-%d", testing.Benchmark(func(b *testing.B) {}).NsPerOp())
	p := &domain.Product{
		Name:            "Integration Code Test POS",
		SKUCode:         sku,
		Category:        "Obat Bebas",
		Unit:            "Pcs",
		ItemsPerUnit:    1,
		StorageLocation: "A1",
		PurchasePrice:   1000,
		SellingPrice:    1500,
		MinStock:        5,
		Status:          "active",
		IsVerified:      true,
	}
	pID, err := productRepo.Create(p)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Batch 1: Expires soon (FEFO)
	b1 := &domain.ProductBatch{
		ProductID:     pID,
		BatchNumber:   "B1-INT-SOON",
		ExpiryDate:    "2026-05-01",
		InitialQty:    10,
		CurrentStock:  10,
		PurchasePrice: 1000,
		SellingPrice:  1500,
		IsVerified:    true,
	}
	b1ID, err := batchRepo.Create(b1)
	if err != nil {
		t.Fatalf("Failed to create batch 1: %v", err)
	}

	// Batch 2: Expires later
	b2 := &domain.ProductBatch{
		ProductID:     pID,
		BatchNumber:   "B2-INT-LATER",
		ExpiryDate:    "2027-01-01",
		InitialQty:    10,
		CurrentStock:  10,
		PurchasePrice: 1000,
		SellingPrice:  1500,
		IsVerified:    true,
	}
	b2ID, err := batchRepo.Create(b2)
	if err != nil {
		t.Fatalf("Failed to create batch 2: %v", err)
	}

	// 2. Process Sale (Need 15: 10 from B1, 5 from B2)
	saleItems := []domain.SaleItem{
		{ProductID: pID, Quantity: 15},
	}

	saleID, err := service.ProcessSale(1, saleItems, "Cash", "Test Integration", "", "", 0)
	if err != nil {
		t.Fatalf("ProcessSale failed: %v", err)
	}

	// Check Stock
	var stock1, stock2 int
	db.QueryRow("SELECT current_stock FROM product_batches WHERE id = ?", b1ID).Scan(&stock1)
	db.QueryRow("SELECT current_stock FROM product_batches WHERE id = ?", b2ID).Scan(&stock2)

	if stock1 != 0 || stock2 != 5 {
		t.Errorf("FEFO failed: Expected B1=0 B2=5, Got B1=%d B2=%d", stock1, stock2)
	}

	// 3. Void Sale
	err = service.VoidSale(saleID, 1)
	if err != nil {
		t.Fatalf("VoidSale failed: %v", err)
	}

	// Check Restoration
	db.QueryRow("SELECT current_stock FROM product_batches WHERE id = ?", b1ID).Scan(&stock1)
	db.QueryRow("SELECT current_stock FROM product_batches WHERE id = ?", b2ID).Scan(&stock2)

	if stock1 != 10 || stock2 != 10 {
		t.Errorf("Restoration failed: Expected both=10, Got B1=%d B2=%d", stock1, stock2)
	}
}

func TestSaleService_DiscountValidation(t *testing.T) {
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
		sub_unit TEXT,
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
		sub_unit TEXT,
		items_per_unit INTEGER
	);
	CREATE TABLE activity_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		action TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	productRepo := repository.NewProductRepository(db)
	batchRepo := repository.NewBatchRepository(db)
	logRepo := repository.NewActivityLogRepository(db)
	saleRepo := repository.NewSaleRepository(db)
	s := NewSaleService(db, saleRepo, productRepo, batchRepo, logRepo)

	// Setup 1 item: 1000 price, 10 stock
	pID, err := productRepo.Create(&domain.Product{Name: "Test", Status: "active", SellingPrice: 1000, Unit: "Pcs", SubUnit: "Pcs", ItemsPerUnit: 1})
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	_, err = batchRepo.Create(&domain.ProductBatch{ProductID: pID, InitialQty: 20, CurrentStock: 10, SellingPrice: 1000, IsVerified: true})
	if err != nil {
		t.Fatalf("Failed to create batch: %v", err)
	}

	tests := []struct {
		name     string
		discount float64
		wantErr  bool
	}{
		{"Accept zero discount", 0, false},
		{"Accept partial discount", 500, false},
		{"Accept full discount", 1000, false},
		{"Reject negative discount", -1, true},
		{"Reject excessive discount", 1001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saleItems := []domain.SaleItem{{ProductID: pID, Quantity: 1}}
			_, err := s.ProcessSale(1, saleItems, "Cash", "Test", "", "", tt.discount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessSale() for %s: error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
