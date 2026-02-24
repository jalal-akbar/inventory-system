package handler

import (
	"encoding/json"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"net/http"
)

type InventoryApiHandler struct {
	BaseHandler
	batchService service.BatchService
}

func NewInventoryApiHandler(base BaseHandler, bService service.BatchService) *InventoryApiHandler {
	return &InventoryApiHandler{
		BaseHandler:  base,
		batchService: bService,
	}
}

func (h *InventoryApiHandler) Adjust(w http.ResponseWriter, r *http.Request) {
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

	var data struct {
		BatchID      int    `json:"batch_id"`
		Quantity     int    `json:"quantity"`
		Unit         string `json:"unit"`
		ItemsPerUnit int    `json:"items_per_unit"`
		Reason       string `json:"reason"`
		Note         string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if data.ItemsPerUnit < 1 {
		data.ItemsPerUnit = 1
	}
	if data.Unit == "" {
		data.Unit = "Pcs"
	}

	if err := h.batchService.AdjustStock(data.BatchID, data.Quantity, data.Unit, data.ItemsPerUnit, data.Reason, data.Note, userID); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "Stock adjusted successfully"})
}
