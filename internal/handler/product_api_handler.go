package handler

import (
	"encoding/json"
	"fmt"
	"inventory-system/internal/domain"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ProductApiHandler struct {
	BaseHandler
	productService service.ProductService
	reportService  service.ReportService
}

func NewProductApiHandler(base BaseHandler, pService service.ProductService, rService service.ReportService) *ProductApiHandler {
	return &ProductApiHandler{
		BaseHandler:    base,
		productService: pService,
		reportService:  rService,
	}
}

func (h *ProductApiHandler) Index(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	filter := r.URL.Query().Get("filter")

	products, err := h.productService.SearchProducts(search, filter)
	if err != nil {
		if r.Header.Get("HX-Request") == "true" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		session := middleware.GetSession(r)
		lang := "en"
		if session != nil {
			lang = session.Lang
		}
		h.RenderStandalone(w, r, "inventory/table_rows", map[string]interface{}{
			"Products": products,
			"Lang":     lang,
		})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"data": products})
}

func (h *ProductApiHandler) Store(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := middleware.GetSession(r)
	if session == nil {
		h.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	userID := session.UserID
	role := session.Role

	var name, sku, category, therapeuticClass, unit, baseUnit, storage, batch, expiry string
	var itemsPerUnit, minStock int
	var purchasePrice, sellingPrice, initialStock float64

	if r.Header.Get("Content-Type") == "application/json" {
		var data struct {
			Name             string  `json:"name"`
			SKUCode          string  `json:"sku_code"`
			Category         string  `json:"category"`
			TherapeuticClass string  `json:"therapeutic_class"`
			Unit             string  `json:"unit"`
			BaseUnit         string  `json:"base_unit"`
			ItemsPerUnit     int     `json:"items_per_unit"`
			StorageLocation  string  `json:"storage_location"`
			PurchasePrice    float64 `json:"purchase_price"`
			SellingPrice     float64 `json:"selling_price"`
			MinStock         int     `json:"min_stock"`
			BatchNumber      string  `json:"batch_number"`
			ExpiryDate       string  `json:"expiry_date"`
			InitialStock     float64 `json:"initial_stock"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		name = data.Name
		sku = data.SKUCode
		category = data.Category
		therapeuticClass = data.TherapeuticClass
		unit = data.Unit
		baseUnit = data.BaseUnit
		itemsPerUnit = data.ItemsPerUnit
		storage = data.StorageLocation
		purchasePrice = data.PurchasePrice
		sellingPrice = data.SellingPrice
		minStock = data.MinStock
		batch = data.BatchNumber
		expiry = data.ExpiryDate
		initialStock = data.InitialStock
	} else {
		name = r.FormValue("name")
		sku = r.FormValue("sku_code")
		category = r.FormValue("category")
		therapeuticClass = r.FormValue("therapeutic_class")
		unit = r.FormValue("unit")
		baseUnit = r.FormValue("base_unit")
		itemsPerUnit, _ = strconv.Atoi(r.FormValue("items_per_unit"))
		storage = r.FormValue("storage_location")
		purchasePrice, _ = strconv.ParseFloat(r.FormValue("purchase_price"), 64)
		sellingPrice, _ = strconv.ParseFloat(r.FormValue("selling_price"), 64)
		minStock, _ = strconv.Atoi(r.FormValue("min_stock"))
		batch = r.FormValue("batch_number")
		expiry = r.FormValue("expiry_date")
		initialStock, _ = strconv.ParseFloat(r.FormValue("initial_stock"), 64)
	}

	if itemsPerUnit < 1 {
		itemsPerUnit = 1
	}

	if !domain.IsValidUnit(unit) {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", `{"toast": {"type": "error", "message": "Invalid unit. Must be one of: `+strings.Join(domain.ValidUnits, ", ")+`"}}`)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid unit. Must be one of: " + strings.Join(domain.ValidUnits, ", ")})
		return
	}

	if !domain.IsValidCategory(category) {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", `{"toast": {"type": "error", "message": "Invalid category. Must be one of: `+strings.Join(domain.ValidCategories, ", ")+`"}}`)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid category. Must be one of: " + strings.Join(domain.ValidCategories, ", ")})
		return
	}

	p := &domain.Product{
		Name:             name,
		SKUCode:          sku,
		Category:         category,
		TherapeuticClass: therapeuticClass,
		Unit:             unit,
		BaseUnit:         baseUnit,
		ItemsPerUnit:     itemsPerUnit,
		StorageLocation:  storage,
		PurchasePrice:    purchasePrice,
		SellingPrice:     sellingPrice,
		MinStock:         minStock,
		Status:           "active",
		IsVerified:       role == "admin",
	}

	id, err := h.productService.CreateProduct(p, batch, expiry, int(initialStock), userID)
	if err != nil {
		log.Printf("Error creating product: %v", err)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", `{"toast": {"type": "error", "message": "Failed to create product: `+err.Error()+`"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", `{"toast": {"type": "success", "message": "Product created successfully"}}`)
		// Redirect to inventory to refresh
		w.Header().Set("HX-Location", "/products")
		return
	}

	h.RespondJSON(w, http.StatusCreated, map[string]interface{}{"status": "success", "id": id, "message": "Product created successfully"})
}

func (h *ProductApiHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := middleware.GetSession(r)
	if session == nil {
		h.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	role := session.Role

	var id, itemsPerUnit, minStock int
	var name, sku, category, therapeuticClass, unit, baseUnit, storage string
	var purchasePrice, sellingPrice float64

	if r.Header.Get("Content-Type") == "application/json" {
		var data struct {
			ID               int     `json:"id"`
			Name             string  `json:"name"`
			SKUCode          string  `json:"sku_code"`
			Category         string  `json:"category"`
			TherapeuticClass string  `json:"therapeutic_class"`
			Unit             string  `json:"unit"`
			BaseUnit         string  `json:"base_unit"`
			ItemsPerUnit     int     `json:"items_per_unit"`
			StorageLocation  string  `json:"storage_location"`
			PurchasePrice    float64 `json:"purchase_price"`
			SellingPrice     float64 `json:"selling_price"`
			MinStock         int     `json:"min_stock"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		id = data.ID
		name = data.Name
		sku = data.SKUCode
		category = data.Category
		therapeuticClass = data.TherapeuticClass
		unit = data.Unit
		baseUnit = data.BaseUnit
		itemsPerUnit = data.ItemsPerUnit
		storage = data.StorageLocation
		purchasePrice = data.PurchasePrice
		sellingPrice = data.SellingPrice
		minStock = data.MinStock
	} else {
		id, _ = strconv.Atoi(r.FormValue("id"))
		name = r.FormValue("name")
		sku = r.FormValue("sku_code")
		idStr := r.FormValue("id")
		if idStr != "" {
			var err error
			id, err = strconv.Atoi(idStr)
			if err != nil {
				h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid value for id"})
				return
			}
		}

		name = r.FormValue("name")
		sku = r.FormValue("sku_code")
		category = r.FormValue("category")
		therapeuticClass = r.FormValue("therapeutic_class")
		unit = r.FormValue("unit")
		baseUnit = r.FormValue("base_unit")

		itemsPerUnitStr := r.FormValue("items_per_unit")
		if itemsPerUnitStr != "" {
			var err error
			itemsPerUnit, err = strconv.Atoi(itemsPerUnitStr)
			if err != nil {
				h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid value for items_per_unit"})
				return
			}
		}

		storage = r.FormValue("storage_location")

		purchasePriceStr := r.FormValue("purchase_price")
		if purchasePriceStr != "" {
			var err error
			purchasePrice, err = strconv.ParseFloat(purchasePriceStr, 64)
			if err != nil {
				h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid value for purchase_price"})
				return
			}
		}

		sellingPriceStr := r.FormValue("selling_price")
		if sellingPriceStr != "" {
			var err error
			sellingPrice, err = strconv.ParseFloat(sellingPriceStr, 64)
			if err != nil {
				h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid value for selling_price"})
				return
			}
		}

		minStockStr := r.FormValue("min_stock")
		if minStockStr != "" {
			var err error
			minStock, err = strconv.Atoi(minStockStr)
			if err != nil {
				h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid value for min_stock"})
				return
			}
		}
	}

	if itemsPerUnit < 1 {
		itemsPerUnit = 1
	}

	if !domain.IsValidUnit(unit) {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", `{"toast": {"type": "error", "message": "Invalid unit. Must be one of: `+strings.Join(domain.ValidUnits, ", ")+`"}}`)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid unit. Must be one of: " + strings.Join(domain.ValidUnits, ", ")})
		return
	}

	p := &domain.Product{
		Name:             name,
		SKUCode:          sku,
		Category:         category,
		TherapeuticClass: therapeuticClass,
		Unit:             unit,
		BaseUnit:         baseUnit,
		ItemsPerUnit:     itemsPerUnit,
		StorageLocation:  storage,
		MinStock:         minStock,
		PurchasePrice:    purchasePrice,
		SellingPrice:     sellingPrice,
	}

	if err := h.productService.UpdateProduct(id, p, role, session.UserID); err != nil {
		log.Printf("Error updating product %d: %v", id, err)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", `{"toast": {"type": "error", "message": "Failed to update product: `+err.Error()+`"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", `{"toast": {"type": "success", "message": "Product updated successfully"}}`)
		w.Header().Set("HX-Location", "/products")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "Product updated successfully"})
}

func (h *ProductApiHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var id int
	if r.Header.Get("Content-Type") == "application/json" {
		var data struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		id = data.ID
	} else {
		id, _ = strconv.Atoi(r.FormValue("id"))
	}

	session := middleware.GetSession(r)
	userID := 0
	if session != nil {
		userID = session.UserID
	}

	if err := h.productService.DeleteProduct(id, userID); err != nil {
		log.Printf("Error deleting product %d: %v", id, err)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", `{"toast": {"type": "error", "message": "Failed to delete product: `+err.Error()+`"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", `{"toast": {"type": "success", "message": "Product deleted successfully"}}`)
		w.Header().Set("HX-Location", "/products")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "Product deleted successfully"})
}

func (h *ProductApiHandler) Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var id int
	if r.Header.Get("Content-Type") == "application/json" {
		var data struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		id = data.ID
	} else {
		id, _ = strconv.Atoi(r.FormValue("id"))
	}

	session := middleware.GetSession(r)
	userID := 0
	if session != nil {
		userID = session.UserID
	}

	if err := h.productService.VerifyProduct(id, userID); err != nil {
		log.Printf("Error verifying product %d: %v", id, err)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", `{"toast": {"type": "error", "message": "Failed to verify product: `+err.Error()+`"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", `{"toast": {"type": "success", "message": "Product verified successfully"}}`)
		w.Header().Set("HX-Location", "/products")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "Product verified successfully"})
}

func (h *ProductApiHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var id int
	if r.Header.Get("Content-Type") == "application/json" {
		var data struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
			return
		}
		id = data.ID
	} else {
		id, _ = strconv.Atoi(r.FormValue("id"))
	}

	session := middleware.GetSession(r)
	userID := 0
	if session != nil {
		userID = session.UserID
	}

	newStatus, err := h.productService.ToggleProductStatus(id, userID)
	if err != nil {
		log.Printf("Error toggling product status %d: %v", id, err)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", `{"toast": {"type": "error", "message": "Failed to toggle status: `+err.Error()+`"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	msg := "Product activated"
	if newStatus == "inactive" {
		msg = "Product deactivated"
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", `{"toast": {"type": "success", "message": "`+msg+`"}}`)
		w.Header().Set("HX-Location", "/products")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": msg, "new_status": newStatus})
}

func (h *ProductApiHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	details, err := h.productService.GetDetails(id)
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": details})
}

func (h *ProductApiHandler) GetPendingCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.productService.GetPendingCount()
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "count": count})
}

func (h *ProductApiHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	results, err := h.productService.QuickSearch(query)
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": results})
}

func (h *ProductApiHandler) GetLedger(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" {
		startDate = "2000-01-01" // Far past
	}
	if endDate == "" {
		endDate = "2099-12-31" // Far future
	}

	startBalance, entries, err := h.reportService.GetProductLedger(id, startDate, endDate)
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "success",
		"start_balance": startBalance,
		"entries":       entries,
	})
}
func (h *ProductApiHandler) BulkStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		IDs    []int  `json:"ids"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	session := middleware.GetSession(r)
	userID := 0
	if session != nil {
		userID = session.UserID
	}

	if err := h.productService.BulkToggleStatus(data.IDs, data.Status, userID); err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "Products updated successfully"})
}

func (h *ProductApiHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		IDs []int `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	session := middleware.GetSession(r)
	userID := 0
	if session != nil {
		userID = session.UserID
	}

	if err := h.productService.BulkDelete(data.IDs, userID); err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "Products deleted successfully"})
}

func (h *ProductApiHandler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	filter := r.URL.Query().Get("filter")

	products, err := h.productService.SearchProducts(search, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f := excelize.NewFile()
	sheet := "Products"
	index, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	// Set Header
	headers := []string{"SKU", "Name", "Category", "Class", "Stock (Pcs)", "Unit", "Purchase Price", "Selling Price", "Expiry", "Status"}
	for i, head := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, head)
	}

	// Set Rows
	for i, p := range products {
		rowIdx := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), p["sku_code"])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIdx), p["name"])
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIdx), p["category"])
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowIdx), p["therapeutic_class"])
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowIdx), p["total_stock"])
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowIdx), p["unit"])
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowIdx), p["purchase_price"])
		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowIdx), p["selling_price"])
		f.SetCellValue(sheet, fmt.Sprintf("I%d", rowIdx), p["nearest_expiry"])
		f.SetCellValue(sheet, fmt.Sprintf("J%d", rowIdx), p["status"])
	}

	f.SetActiveSheet(index)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=inventory_report.xlsx")

	if err := f.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
