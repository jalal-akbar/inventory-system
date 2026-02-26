package service

import (
	"database/sql"
	"errors"
	"fmt"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

type SaleService interface {
	ProcessSale(userID int, items []domain.SaleItem, paymentMethod, customerName, doctorName, prescriptionNumber string, discount float64) (int, error)
	VoidSale(saleID, adminUserID int) error
	RequestVoid(saleID, userID int, reason string) error
	GetSaleDetails(saleID int) (map[string]interface{}, error)
}

type saleService struct {
	db          *sql.DB
	saleRepo    repository.SaleRepository
	productRepo repository.ProductRepository
	batchRepo   repository.ProductBatchRepository
	logRepo     repository.ActivityLogRepository
}

func NewSaleService(db *sql.DB, sRepo repository.SaleRepository, pRepo repository.ProductRepository, bRepo repository.ProductBatchRepository, lRepo repository.ActivityLogRepository) SaleService {
	return &saleService{
		db:          db,
		saleRepo:    sRepo,
		productRepo: pRepo,
		batchRepo:   bRepo,
		logRepo:     lRepo,
	}
}

func (s *saleService) ProcessSale(userID int, items []domain.SaleItem, paymentMethod, customerName, doctorName, prescriptionNumber string, discount float64) (int, error) {
	// Normalize payment method
	switch paymentMethod {
	case "cash", "Cash":
		paymentMethod = "Cash"
	case "transfer", "Transfer":
		paymentMethod = "Transfer"
	default:
		return 0, fmt.Errorf("invalid payment method: %s. Supported: Cash, Transfer", paymentMethod)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Wrap repositories with transaction
	productRepo := s.productRepo.WithTx(tx)
	batchRepo := s.batchRepo.WithTx(tx)
	saleRepo := s.saleRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	var totalAmount float64
	var totalProfit float64
	var itemsToInsert []domain.SaleItem

	for _, item := range items {
		if item.Quantity <= 0 {
			continue
		}

		p, err := productRepo.FindByID(item.ProductID)
		if err != nil || p == nil || p.Status != "active" {
			return 0, fmt.Errorf("product ID %d not found or inactive", item.ProductID)
		}

		// psikotropika validation
		if p.Category == "Psikotropika" {
			if customerName == "" || doctorName == "" || prescriptionNumber == "" {
				return 0, fmt.Errorf("transaksi item Psikotropika (%s) wajib mengisi Nama Customer, Nama Dokter, dan Nomor Resep", p.Name)
			}
		}

		// Treat quantity as PCS directly (Opsi A)
		effectiveQtyNeeded := item.Quantity

		batches, err := batchRepo.GetAvailableForProduct(item.ProductID)
		if err != nil {
			return 0, err
		}

		remainingNeeded := effectiveQtyNeeded
		for _, batch := range batches {
			if remainingNeeded <= 0 {
				break
			}

			qtyToTake := remainingNeeded
			if batch.CurrentStock < qtyToTake {
				qtyToTake = batch.CurrentStock
			}

			if err := batchRepo.UpdateStock(batch.ID, -qtyToTake); err != nil {
				return 0, err
			}

			// Subtotal is correctly calculated based on PCS taken and unit prices (which are per PCS)
			itemSubtotal := float64(qtyToTake) * batch.SellingPrice
			itemCost := float64(qtyToTake) * batch.PurchasePrice

			itemsToInsert = append(itemsToInsert, domain.SaleItem{
				ProductID:    item.ProductID,
				BatchID:      batch.ID,
				Quantity:     qtyToTake,
				Price:        batch.SellingPrice,
				Subtotal:     itemSubtotal,
				SaleUnit:     p.Unit,
				ItemsPerUnit: p.ItemsPerUnit,
			})

			totalAmount += itemSubtotal
			totalProfit += (itemSubtotal - itemCost)
			remainingNeeded -= qtyToTake
		}

		if remainingNeeded > 0 {
			return 0, fmt.Errorf("insufficient stock for product: %s", p.Name)
		}
	}

	if len(itemsToInsert) == 0 {
		return 0, errors.New("cannot process empty sale")
	}

	if discount < 0 {
		return 0, errors.New("discount cannot be negative")
	}

	if discount > totalAmount {
		return 0, fmt.Errorf("discount (%.2f) cannot exceed total amount (%.2f)", discount, totalAmount)
	}

	finalTotal := totalAmount - discount
	finalProfit := totalProfit - discount

	sale := &domain.Sale{
		UserID:        userID,
		TotalAmount:   finalTotal,
		Profit:        finalProfit,
		Discount:      discount,
		PaymentMethod: paymentMethod,
		Status:        "active",
	}

	if customerName != "" {
		sale.CustomerName = &customerName
	}
	if doctorName != "" {
		sale.DoctorName = &doctorName
	}
	if prescriptionNumber != "" {
		sale.PrescriptionNumber = &prescriptionNumber
	}

	saleID, err := saleRepo.CreateSale(sale)
	if err != nil {
		return 0, err
	}

	for _, i := range itemsToInsert {
		i.SaleID = saleID
		if err := saleRepo.CreateSaleItem(&i); err != nil {
			return 0, err
		}
	}

	if err := logRepo.Log(userID, fmt.Sprintf("Processed sale #%d - Total: Rp %.2f", saleID, finalTotal)); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return saleID, nil
}

func (s *saleService) VoidSale(saleID, adminUserID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Wrap repositories
	saleRepo := s.saleRepo.WithTx(tx)
	batchRepo := s.batchRepo.WithTx(tx)
	logRepo := s.logRepo.WithTx(tx)

	sale, err := saleRepo.FindByID(saleID)
	if err != nil {
		return err
	}
	if sale == nil {
		return fmt.Errorf("sale #%d not found", saleID)
	}

	if sale.Status == "void" {
		return fmt.Errorf("sale #%d is already voided", saleID)
	}

	if err := saleRepo.SetStatus(saleID, "void"); err != nil {
		return err
	}

	items, err := saleRepo.GetSaleItems(saleID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := batchRepo.UpdateStock(item.BatchID, item.Quantity); err != nil {
			return err
		}
	}

	if err := logRepo.Log(adminUserID, fmt.Sprintf("Voided sale #%d - Amount: Rp %.2f", saleID, sale.TotalAmount)); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *saleService) RequestVoid(saleID, userID int, reason string) error {
	return s.saleRepo.CreateVoidRequest(saleID, reason, userID)
}

func (s *saleService) GetSaleDetails(saleID int) (map[string]interface{}, error) {
	sale, err := s.saleRepo.FindByID(saleID)
	if err != nil {
		return nil, err
	}
	if sale == nil {
		return nil, errors.New("sale not found")
	}

	items, err := s.saleRepo.GetSaleItems(saleID)
	if err != nil {
		return nil, err
	}

	var itemsDetailed []map[string]interface{}
	for _, item := range items {
		p, _ := s.productRepo.FindByID(item.ProductID)
		batch, _ := s.batchRepo.FindByID(item.BatchID)

		name := "Unknown Product"
		if p != nil {
			name = p.Name
		}
		batchNum := "Unknown"
		if batch != nil {
			batchNum = batch.BatchNumber
		}

		itemsDetailed = append(itemsDetailed, map[string]interface{}{
			"id":             item.ID,
			"product_id":     item.ProductID,
			"product_name":   name,
			"batch_id":       item.BatchID,
			"batch_number":   batchNum,
			"quantity":       item.Quantity,
			"price":          item.Price,
			"subtotal":       item.Subtotal,
			"sale_unit":      item.SaleUnit,
			"items_per_unit": item.ItemsPerUnit,
		})
	}

	return map[string]interface{}{
		"sale":  sale,
		"items": itemsDetailed,
	}, nil
}
