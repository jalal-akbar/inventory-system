package handler

import (
	"encoding/json"
	"inventory-system/internal/domain"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ProductApiHandler struct {
	BaseHandler
	productService service.ProductService
}

func NewProductApiHandler(base BaseHandler, pService service.ProductService) *ProductApiHandler {
	return &ProductApiHandler{
		BaseHandler:    base,
		productService: pService,
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

	var name, sku, category, therapeuticClass, unit, subUnit, storage, batch, expiry string
	var itemsPerUnit, minStock int
	var purchasePrice, sellingPrice, initialStock float64

	if r.Header.Get("Content-Type") == "application/json" {
		var data struct {
			Name             string  `json:"name"`
			SKUCode          string  `json:"sku_code"`
			Category         string  `json:"category"`
			TherapeuticClass string  `json:"therapeutic_class"`
			Unit             string  `json:"unit"`
			SubUnit          string  `json:"sub_unit"`
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
		subUnit = data.SubUnit
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
		subUnit = r.FormValue("sub_unit")
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
		SubUnit:          subUnit,
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
	var name, sku, category, therapeuticClass, unit, subUnit, storage string
	var purchasePrice, sellingPrice float64

	if r.Header.Get("Content-Type") == "application/json" {
		var data struct {
			ID               int     `json:"id"`
			Name             string  `json:"name"`
			SKUCode          string  `json:"sku_code"`
			Category         string  `json:"category"`
			TherapeuticClass string  `json:"therapeutic_class"`
			Unit             string  `json:"unit"`
			SubUnit          string  `json:"sub_unit"`
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
		subUnit = data.SubUnit
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
		subUnit = r.FormValue("sub_unit")

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
		SubUnit:          subUnit,
		ItemsPerUnit:     itemsPerUnit,
		StorageLocation:  storage,
		MinStock:         minStock,
		PurchasePrice:    purchasePrice,
		SellingPrice:     sellingPrice,
	}

	if err := h.productService.UpdateProduct(id, p, role); err != nil {
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

	if err := h.productService.DeleteProduct(id); err != nil {
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

	if err := h.productService.VerifyProduct(id); err != nil {
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
