package handler

import (
	"encoding/json"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"net/http"
	"strconv"
)

type ReturnHandler struct {
	BaseHandler
	ReturnService service.ReturnService
	SaleService   service.SaleService
}

func NewReturnHandler(base BaseHandler, rService service.ReturnService, sService service.SaleService) *ReturnHandler {
	return &ReturnHandler{
		BaseHandler:   base,
		ReturnService: rService,
		SaleService:   sService,
	}
}

func (h *ReturnHandler) ShowForm(w http.ResponseWriter, r *http.Request) {
	saleIDStr := r.URL.Query().Get("sale_id")
	saleID, _ := strconv.Atoi(saleIDStr)

	if saleID == 0 {
		http.Redirect(w, r, "/reports/history", http.StatusSeeOther)
		return
	}

	// For the form, we need the sale items.
	// The SaleService doesn't have a GetSaleByID with items method exposed in interface yet.
	// But we can use the repository if we had it, or just use the service if we add it.
	// Actually, the ReturnService needs to be able to show what's returnable.

	// Let's assume we have a way to get sale details.
	// For now, I'll render the template and let the template handle the data if possible,
	// or I'll just fetch what's needed here.

	// I'll need a way to get sale items. I'll check if I can add a method to SaleService.
	// Or I can just fetch them via repository directly if I inject it.
	// But let's stick to Service layer.

	// I'll add GetSaleDetails to SaleService or just use a helper here if I had access to Repo.
	// Wait, I can just use the ReturnService to get what's returnable if I implement a method there.

	// Actually, let's keep it simple: Render the page, and the page can fetch details via AJAX/HTMX
	// or we pass it in. Passing it in is better for initial load.

	h.Render(w, r, "sales/return", map[string]interface{}{
		"PageTitle":   "Process Return",
		"CurrentPage": "returns",
		"SaleID":      saleID,
	})
}

func (h *ReturnHandler) Store(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := middleware.GetSession(r)
	if session == nil {
		h.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		SaleID int    `json:"sale_id"`
		Reason string `json:"reason"`
		Items  []struct {
			SaleItemID int    `json:"sale_item_id"`
			Quantity   int    `json:"quantity"`
			Condition  string `json:"condition"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var serviceItems []service.ReturnInputItem
	for _, it := range req.Items {
		if it.Quantity > 0 {
			serviceItems = append(serviceItems, service.ReturnInputItem{
				SaleItemID: it.SaleItemID,
				Quantity:   it.Quantity,
				Condition:  it.Condition,
			})
		}
	}

	returnID, err := h.ReturnService.ProcessReturn(session.UserID, req.SaleID, req.Reason, serviceItems)
	if err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"return_id": returnID,
		"message":   "Return processed successfully",
	})
}
