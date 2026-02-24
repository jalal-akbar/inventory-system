package handler

import (
	"inventory-system/internal/i18n"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"log"
	"net/http"
	"strconv"
)

type ProductHandler struct {
	BaseHandler
	productService service.ProductService
	batchService   service.BatchService
}

func NewProductHandler(base BaseHandler, pService service.ProductService, bService service.BatchService) *ProductHandler {
	return &ProductHandler{
		BaseHandler:    base,
		productService: pService,
		batchService:   bService,
	}
}

func (h *ProductHandler) Index(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	filter := r.URL.Query().Get("filter")
	products, _ := h.productService.SearchProducts(search, filter)

	lang := h.GetLang(r)
	h.Render(w, r, "inventory/index", map[string]interface{}{
		"pageTitle":    i18n.T("inventory", lang),
		"browserTitle": i18n.T("inventory", lang) + " - " + "Inventory System",
		"CurrentPage":  "inventory",
		"Products":     products,
		"Filters": map[string]string{
			"search": search,
			"filter": filter,
		},
	})
}

func (h *ProductHandler) InventoryCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.ProcessInventoryCheck(w, r)
		return
	}

	products, _ := h.productService.SearchWithAllBatches("", "")

	lang := h.GetLang(r)
	h.Render(w, r, "inventory/check", map[string]interface{}{
		"pageTitle":    i18n.T("inventory_status", lang),
		"browserTitle": i18n.T("inventory_status", lang) + " - " + "Inventory System",
		"CurrentPage":  "stock_check",
		"Products":     products,
	})
}

func (h *ProductHandler) ProcessInventoryCheck(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	session := middleware.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := session.UserID

	// Form expected to have fields like actual_stock_{batch_id}
	for key, values := range r.PostForm {
		if len(values) == 0 || values[0] == "" {
			continue
		}

		// Look for keys like actual_stock_123
		if len(key) > 13 && key[:13] == "actual_stock_" {
			batchIDStr := key[13:]
			batchID, _ := strconv.Atoi(batchIDStr)
			actualStock, _ := strconv.Atoi(values[0])

			// Extract unit and itemsPerUnit from corresponding form fields
			unit := r.FormValue("unit_" + batchIDStr)
			if unit == "" {
				unit = "Pcs"
			}
			itemsPerUnitStr := r.FormValue("items_per_unit_" + batchIDStr)
			itemsPerUnit, _ := strconv.Atoi(itemsPerUnitStr)
			if itemsPerUnit < 1 {
				itemsPerUnit = 1
			}

			// Perform adjustment with unit context
			if err := h.batchService.PerformInventoryCheck(batchID, actualStock, unit, itemsPerUnit, "Manual Inventory Check", userID); err != nil {
				// Log error but continue
				log.Printf("Error processing inventory check for batch %d: %v", batchID, err)
			}
		}
	}

	http.Redirect(w, r, "/inventory/check?message=Inventory+adjustments+saved+successfully", http.StatusSeeOther)
}

func (h *ProductHandler) PrintLabel(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	product, err := h.productService.GetDetails(id)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	lang := h.GetLang(r)
	h.Render(w, r, "products/print_label", map[string]interface{}{
		"pageTitle":    i18n.T("print_label", lang),
		"browserTitle": i18n.T("print_label", lang) + " - " + "Inventory System",
		"Product":      product["product"],
	})
}
