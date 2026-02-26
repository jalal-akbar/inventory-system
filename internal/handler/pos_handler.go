package handler

import (
	"encoding/json"
	"inventory-system/internal/domain"
	"inventory-system/internal/i18n"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"net/http"
)

type PosHandler struct {
	BaseHandler
	productService service.ProductService
	saleService    service.SaleService
}

func NewPosHandler(base BaseHandler, pService service.ProductService, sService service.SaleService) *PosHandler {
	return &PosHandler{
		BaseHandler:    base,
		productService: pService,
		saleService:    sService,
	}
}

func (h *PosHandler) Index(w http.ResponseWriter, r *http.Request) {
	// Fetch initial products (limit 10 best sellers)
	products, err := h.productService.GetBestSellingProducts(10)
	if err != nil {
		products = []map[string]interface{}{}
	}

	categories := domain.ValidTherapeuticClasses

	lang := h.GetLang(r)
	h.RenderStandalone(w, r, "pos/index", map[string]interface{}{
		"pageTitle":    i18n.T("point_of_sale", lang),
		"browserTitle": i18n.T("point_of_sale", lang) + " - " + "Inventory System",
		"CurrentPage":  "pos",
		"Products":     products,
		"Categories":   categories,
	})
}

func (h *PosHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")

	products, err := h.productService.SearchProducts(query, category)
	if err != nil {
		products = []map[string]interface{}{}
	}

	h.RenderStandalone(w, r, "pos/_search_results", map[string]interface{}{
		"Products": products,
	})
}

func (h *PosHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := middleware.GetSession(r)
	if session == nil {
		h.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var data struct {
		Items []struct {
			ID  int `json:"id"`
			Qty int `json:"qty"`
		} `json:"items"`
		PaymentMethod string  `json:"payment_method"`
		CustomerName  string  `json:"customer_name"`
		Discount      float64 `json:"discount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var saleItems []domain.SaleItem
	for _, item := range data.Items {
		saleItems = append(saleItems, domain.SaleItem{
			ProductID: item.ID,
			Quantity:  item.Qty,
		})
	}

	saleID, err := h.saleService.ProcessSale(
		session.UserID,
		saleItems,
		data.PaymentMethod,
		data.CustomerName,
		"", // doctorName (optional for now)
		"", // prescriptionNumber (optional for now)
		data.Discount,
	)

	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"sale_id": saleID,
		"message": "Transaction completed successfully",
	})
}
