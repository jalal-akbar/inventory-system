package service

import (
	"database/sql"
	"errors"
	"fmt"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

type ReturnService interface {
	ProcessReturn(userID int, saleID int, reason string, items []ReturnInputItem) (int, error)
	GetReturnDetails(returnID int) (*domain.ReturnDetail, error)
	GetReturnsBySaleID(saleID int) ([]domain.Return, error)
}

type ReturnInputItem struct {
	SaleItemID int
	Quantity   int
	Condition  string // good, damaged
}

type returnService struct {
	db         *sql.DB
	returnRepo repository.ReturnRepository
	saleRepo   repository.SaleRepository
	batchRepo  repository.ProductBatchRepository
	logRepo    repository.ActivityLogRepository
}

func NewReturnService(db *sql.DB, rRepo repository.ReturnRepository, sRepo repository.SaleRepository, bRepo repository.ProductBatchRepository, lRepo repository.ActivityLogRepository) ReturnService {
	return &returnService{
		db:         db,
		returnRepo: rRepo,
		saleRepo:   sRepo,
		batchRepo:  bRepo,
		logRepo:    lRepo,
	}
}

func (s *returnService) ProcessReturn(userID int, saleID int, reason string, items []ReturnInputItem) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Wrap repos
	returnRepo := s.returnRepo.WithTx(tx)
	saleRepo := s.saleRepo.WithTx(tx)
	batchRepo := s.batchRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	// 1. Validate Sale
	sale, err := saleRepo.FindByID(saleID)
	if err != nil {
		return 0, err
	}
	if sale == nil {
		return 0, fmt.Errorf("sale #%d not found", saleID)
	}
	if sale.Status != "active" {
		return 0, fmt.Errorf("cannot return items from a %s sale", sale.Status)
	}

	// 2. Fetch Sale Items for price calculation and validation
	saleItems, err := saleRepo.GetSaleItems(saleID)
	if err != nil {
		return 0, err
	}
	saleItemsMap := make(map[int]domain.SaleItem)
	for _, si := range saleItems {
		saleItemsMap[si.ID] = si
	}

	var totalRefund float64
	var returnItems []domain.ReturnItem

	for _, inputItem := range items {
		if inputItem.Quantity <= 0 {
			continue
		}

		si, ok := saleItemsMap[inputItem.SaleItemID]
		if !ok {
			return 0, fmt.Errorf("sale item ID %d not found in sale #%d", inputItem.SaleItemID, saleID)
		}

		// Check already returned quantity
		returnedQty, err := returnRepo.GetReturnedQtyBySaleItemID(inputItem.SaleItemID)
		if err != nil {
			return 0, err
		}

		if returnedQty+inputItem.Quantity > si.Quantity {
			return 0, fmt.Errorf("cannot return %d units of product ID %d; only %d units remaining from original sale",
				inputItem.Quantity, si.ProductID, si.Quantity-returnedQty)
		}

		refundAmount := float64(inputItem.Quantity) * si.Price
		totalRefund += refundAmount

		returnItems = append(returnItems, domain.ReturnItem{
			SaleItemID:      inputItem.SaleItemID,
			Quantity:        inputItem.Quantity,
			RefundAmount:    refundAmount,
			ConditionStatus: inputItem.Condition,
		})

		// 3. Update Stock if condition is good
		if inputItem.Condition == "good" {
			if err := batchRepo.UpdateStock(si.BatchID, inputItem.Quantity); err != nil {
				return 0, fmt.Errorf("failed to update stock for batch ID %d: %v", si.BatchID, err)
			}
		}
	}

	if len(returnItems) == 0 {
		return 0, errors.New("no items to return")
	}

	// 4. Create Return Header
	ret := &domain.Return{
		SaleID:      saleID,
		UserID:      userID,
		TotalRefund: totalRefund,
	}
	if reason != "" {
		ret.Reason = &reason
	}

	returnID, err := returnRepo.CreateReturn(ret)
	if err != nil {
		return 0, err
	}

	// 5. Create Return Items
	for _, ri := range returnItems {
		ri.ReturnID = returnID
		if err := returnRepo.CreateReturnItem(&ri); err != nil {
			return 0, err
		}
	}

	// 6. Log Activity
	if err := logRepo.Log(userID, fmt.Sprintf("Processed Return #%d for Sale #%d - Total Refund: Rp %.0f", returnID, saleID, totalRefund)); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return returnID, nil
}

func (s *returnService) GetReturnDetails(returnID int) (*domain.ReturnDetail, error) {
	return s.returnRepo.GetReturnWithDetails(returnID)
}

func (s *returnService) GetReturnsBySaleID(saleID int) ([]domain.Return, error) {
	return s.returnRepo.GetReturnsBySaleID(saleID)
}
