package service

import (
	"database/sql"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

type ProductService interface {
	CreateProduct(p *domain.Product, initialBatchNumber, expiryDate string, initialQty int, requestedBy int) (int, error)
	UpdateProduct(id int, p *domain.Product, userRole string, userID int) error
	DeleteProduct(id int, userID int) error
	VerifyProduct(id int, userID int) error
	ToggleProductStatus(id int, userID int) (string, error)
	SearchProducts(search, filter string) ([]map[string]interface{}, error)
	QuickSearch(query string) ([]map[string]interface{}, error)
	GetDetails(id int) (map[string]interface{}, error)
	GetPendingCount() (int, error)
	GetLowStock() ([]map[string]interface{}, error)
	GetRecentProducts(limit int) ([]map[string]interface{}, error)
	GetBestSellingProducts(limit int) ([]map[string]interface{}, error)
	SearchWithAllBatches(search, filter string) ([]map[string]interface{}, error)
	BulkToggleStatus(ids []int, targetStatus string, userID int) error
	BulkDelete(ids []int, userID int) error
}

type productService struct {
	db          *sql.DB
	productRepo repository.ProductRepository
	batchRepo   repository.ProductBatchRepository
	entryRepo   repository.StockEntryRepository
	logRepo     repository.ActivityLogRepository
}

func NewProductService(db *sql.DB, pRepo repository.ProductRepository, bRepo repository.ProductBatchRepository, eRepo repository.StockEntryRepository, lRepo repository.ActivityLogRepository) ProductService {
	return &productService{
		db:          db,
		productRepo: pRepo,
		batchRepo:   bRepo,
		entryRepo:   eRepo,
		logRepo:     lRepo,
	}
}

func (s *productService) CreateProduct(p *domain.Product, batchNum, expiry string, qty int, requestedBy int) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Wrap repos with transaction
	productRepo := s.productRepo.WithTx(tx)
	batchRepo := s.batchRepo.WithTx(tx)
	entryRepo := s.entryRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	// 1. Create Product
	pID, err := productRepo.Create(p)
	if err != nil {
		return 0, err
	}

	// 2. Create Initial Batch
	batch := &domain.ProductBatch{
		ProductID:     pID,
		BatchNumber:   batchNum,
		ExpiryDate:    expiry,
		InitialQty:    qty,
		CurrentStock:  qty * p.ItemsPerUnit,
		PurchasePrice: p.PurchasePrice,
		SellingPrice:  p.SellingPrice,
		IsVerified:    p.IsVerified,
	}
	bID, err := batchRepo.Create(batch)
	if err != nil {
		return 0, err
	}

	// 3. Create Stock Entry
	entry := &domain.StockEntry{
		ProductID:   pID,
		BatchID:     bID,
		Quantity:    qty,
		Status:      "approved",
		IsVerified:  p.IsVerified,
		RequestedBy: requestedBy,
	}
	if err := entryRepo.Create(entry); err != nil {
		return 0, err
	}

	// 4. Activity Log
	if err := logRepo.Log(requestedBy, "Created product: "+p.Name); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return pID, nil
}

func (s *productService) UpdateProduct(id int, p *domain.Product, role string, userID int) error {
	isAdmin := role == "admin"
	if err := s.productRepo.UpdateFull(id, p, isAdmin); err != nil {
		return err
	}
	return s.logRepo.Log(userID, "Updated product: "+p.Name)
}

func (s *productService) DeleteProduct(id int, userID int) error {
	p, err := s.productRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.SoftDelete(id); err != nil {
		return err
	}

	return s.logRepo.Log(userID, "Inactivated product: "+p.Name)
}

func (s *productService) VerifyProduct(id int, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Wrap repos
	productRepo := s.productRepo.WithTx(tx)
	batchRepo := s.batchRepo.WithTx(tx)
	entryRepo := s.entryRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	p, err := productRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := productRepo.Verify(id); err != nil {
		return err
	}
	if err := batchRepo.VerifyByProduct(id); err != nil {
		return err
	}
	if err := entryRepo.VerifyByProduct(id); err != nil {
		return err
	}

	if err := logRepo.Log(userID, "Verified product: "+p.Name); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *productService) ToggleProductStatus(id int, userID int) (string, error) {
	p, err := s.productRepo.FindByID(id)
	if err != nil {
		return "", err
	}

	newStatus := "active"
	isVerified := 1
	if p.Status == "active" {
		newStatus = "inactive"
		isVerified = 0
	}

	if err := s.productRepo.UpdateStatus(id, newStatus, isVerified); err != nil {
		return "", err
	}

	actionTxt := "Activated"
	if newStatus == "inactive" {
		actionTxt = "Inactivated"
	}

	if err := s.logRepo.Log(userID, actionTxt+" product: "+p.Name); err != nil {
		return "", err
	}

	return newStatus, nil
}

func (s *productService) SearchProducts(search, filter string) ([]map[string]interface{}, error) {
	return s.productRepo.SearchWithStock(search, filter)
}

func (s *productService) QuickSearch(query string) ([]map[string]interface{}, error) {
	return s.productRepo.QuickSearch(query)
}

func (s *productService) GetDetails(id int) (map[string]interface{}, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	batches, err := s.batchRepo.FindByProduct(id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"product": product,
		"batches": batches,
	}, nil
}

func (s *productService) GetPendingCount() (int, error) {
	return s.productRepo.GetPendingCount()
}

func (s *productService) GetLowStock() ([]map[string]interface{}, error) {
	return s.productRepo.GetLowStock()
}

func (s *productService) GetRecentProducts(limit int) ([]map[string]interface{}, error) {
	return s.productRepo.GetRecent(limit)
}

func (s *productService) GetBestSellingProducts(limit int) ([]map[string]interface{}, error) {
	return s.productRepo.GetBestSellers(limit)
}

func (s *productService) SearchWithAllBatches(search, filter string) ([]map[string]interface{}, error) {
	return s.productRepo.SearchWithAllBatches(search, filter)
}
func (s *productService) BulkToggleStatus(ids []int, targetStatus string, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	productRepo := s.productRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	isVerified := 1
	if targetStatus == "inactive" {
		isVerified = 0
	}

	for _, id := range ids {
		p, err := productRepo.FindByID(id)
		if err != nil {
			continue
		}
		if err := productRepo.UpdateStatus(id, targetStatus, isVerified); err != nil {
			return err
		}
		logRepo.Log(userID, "Bulk "+targetStatus+": "+p.Name)
	}

	return tx.Commit()
}

func (s *productService) BulkDelete(ids []int, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	productRepo := s.productRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	for _, id := range ids {
		p, err := productRepo.FindByID(id)
		if err != nil {
			continue
		}
		if err := productRepo.SoftDelete(id); err != nil {
			return err
		}
		logRepo.Log(userID, "Bulk Deleted (Inactive): "+p.Name)
	}

	return tx.Commit()
}
