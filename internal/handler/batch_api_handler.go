package handler

import (
	"encoding/json"
	"inventory-system/internal/domain"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"net/http"
	"strconv"
)

type BatchApiHandler struct {
	BaseHandler
	batchService service.BatchService
}

func NewBatchApiHandler(base BaseHandler, bService service.BatchService) *BatchApiHandler {
	return &BatchApiHandler{
		BaseHandler:  base,
		batchService: bService,
	}
}

func (h *BatchApiHandler) Index(w http.ResponseWriter, r *http.Request) {
	productID, _ := strconv.Atoi(r.URL.Query().Get("product_id"))
	batches, err := h.batchService.GetBatchesByProduct(productID)
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "data": batches})
}

func (h *BatchApiHandler) Store(w http.ResponseWriter, r *http.Request) {
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

	var data struct {
		ProductID     int     `json:"product_id"`
		BatchNumber   string  `json:"batch_number"`
		ExpiryDate    string  `json:"expiry_date"`
		PackageCount  float64 `json:"package_count"`
		ItemsPerUnit  int     `json:"items_per_unit"`
		PurchasePrice float64 `json:"purchase_price"`
		SellingPrice  float64 `json:"selling_price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	b := &domain.ProductBatch{
		ProductID:     data.ProductID,
		BatchNumber:   data.BatchNumber,
		ExpiryDate:    data.ExpiryDate,
		InitialQty:    int(data.PackageCount),
		CurrentStock:  int(data.PackageCount) * data.ItemsPerUnit,
		PurchasePrice: data.PurchasePrice,
		SellingPrice:  data.SellingPrice,
		IsVerified:    role == "admin",
	}

	id, err := h.batchService.AddBatch(b, userID)
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusCreated, map[string]interface{}{"status": "success", "id": id, "message": "Batch added successfully"})
}
