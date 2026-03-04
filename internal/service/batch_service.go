package service

import (
	"database/sql"
	"errors"
	"fmt"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

type BatchService interface {
	AddBatch(b *domain.ProductBatch, requestedBy int) (int, error)
	GetBatchesByProduct(productID int) ([]domain.ProductBatch, error)
	AdjustStock(batchID int, qtyToRemove int, unit string, itemsPerUnit int, reason, note string, userID int) error
	PerformInventoryCheck(batchID int, actualStock int, unit string, itemsPerUnit int, notes string, userID int) error
	GetExpiringBatches(days int) ([]map[string]interface{}, error)
	GetExpiringCount(days int) (int, error)
}

type batchService struct {
	db        *sql.DB
	batchRepo repository.ProductBatchRepository
	entryRepo repository.StockEntryRepository
	logRepo   repository.ActivityLogRepository
}

func NewBatchService(db *sql.DB, bRepo repository.ProductBatchRepository, eRepo repository.StockEntryRepository, lRepo repository.ActivityLogRepository) BatchService {
	return &batchService{
		db:        db,
		batchRepo: bRepo,
		entryRepo: eRepo,
		logRepo:   lRepo,
	}
}

func (s *batchService) AddBatch(b *domain.ProductBatch, requestedBy int) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Wrap repos
	batchRepo := s.batchRepo.WithTx(tx)
	entryRepo := s.entryRepo.WithTx(tx)

	status, err := batchRepo.GetProductStatus(b.ProductID)
	if err != nil {
		return 0, err
	}
	if status != "active" {
		return 0, errors.New("cannot add batch to an inactive product")
	}

	bID, err := batchRepo.Create(b)
	if err != nil {
		return 0, err
	}

	entry := &domain.StockEntry{
		ProductID:   b.ProductID,
		BatchID:     bID,
		Quantity:    b.CurrentStock,
		Status:      "approved",
		IsVerified:  b.IsVerified,
		RequestedBy: requestedBy,
	}
	if err := entryRepo.Create(entry); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return bID, nil
}

func (s *batchService) GetBatchesByProduct(productID int) ([]domain.ProductBatch, error) {
	return s.batchRepo.FindByProduct(productID)
}

func (s *batchService) AdjustStock(batchID int, qtyToRemove int, unit string, itemsPerUnit int, reason, note string, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Wrap repos
	batchRepo := s.batchRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	// 1. Get Batch Info
	data, err := batchRepo.GetWithProduct(batchID)
	if err != nil {
		return err
	}
	batch := data["batch"].(domain.ProductBatch)
	productName := data["product_name"].(string)
	productStatus := data["product_status"].(string)

	if productStatus != "active" {
		return errors.New("cannot adjust stock for an inactive product")
	}

	effectiveQty := qtyToRemove * itemsPerUnit
	if batch.CurrentStock < effectiveQty {
		return errors.New("adjustment quantity exceeds available stock")
	}

	// 2. Update Stock
	if err := batchRepo.UpdateStock(batchID, -effectiveQty); err != nil {
		return err
	}

	// 3. Log Activity
	displayQty := fmt.Sprintf("%d pcs", effectiveQty)
	if unit != "" && unit != "Pcs" {
		displayQty = fmt.Sprintf("%d %s (%d pcs)", qtyToRemove, unit, effectiveQty)
	}
	logMsg := fmt.Sprintf("Stock Adjustment: %s (Batch: %s) - Removed %s. Reason: %s", productName, batch.BatchNumber, displayQty, reason)
	if note != "" {
		logMsg += " (Note: " + note + ")"
	}
	if err := logRepo.Log(userID, logMsg); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *batchService) PerformInventoryCheck(batchID int, actualStock int, unit string, itemsPerUnit int, notes string, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Wrap repos
	batchRepo := s.batchRepo.WithTx(tx)
	entryRepo := s.entryRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	// 1. Get Batch Info
	data, err := batchRepo.GetWithProduct(batchID)
	if err != nil {
		return err
	}
	batch := data["batch"].(domain.ProductBatch)
	productName := data["product_name"].(string)
	productStatus := data["product_status"].(string)

	if productStatus != "active" {
		return errors.New("cannot perform inventory check on an inactive product")
	}

	actualStockPCS := actualStock * itemsPerUnit
	difference := actualStockPCS - batch.CurrentStock
	if difference == 0 {
		return nil // No change needed
	}

	// 2. Set Absolute Stock
	if err := batchRepo.SetStock(batchID, actualStockPCS); err != nil {
		return err
	}

	// 3. Create Stock Entry (Adjustment record)
	entry := &domain.StockEntry{
		ProductID:   batch.ProductID,
		BatchID:     batchID,
		Quantity:    difference,
		Status:      "approved",
		IsVerified:  true,
		RequestedBy: userID,
	}
	if err := entryRepo.Create(entry); err != nil {
		return err
	}

	// 4. Log Activity
	displayActual := fmt.Sprintf("%d pcs", actualStockPCS)
	if unit != "" && unit != "Pcs" {
		displayActual = fmt.Sprintf("%d %s (%d pcs)", actualStock, unit, actualStockPCS)
	}
	logMsg := fmt.Sprintf("Inventory Check: %s (Batch: %s) adjusted from %d pcs to %s.", productName, batch.BatchNumber, batch.CurrentStock, displayActual)
	if notes != "" {
		logMsg += " Notes: " + notes
	}
	if err := logRepo.Log(userID, logMsg); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *batchService) GetExpiringBatches(days int) ([]map[string]interface{}, error) {
	return s.batchRepo.GetExpiringBatches(days)
}

func (s *batchService) GetExpiringCount(days int) (int, error) {
	return s.batchRepo.GetExpiringCount(days)
}
