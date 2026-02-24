package handler

import (
	"encoding/json"
	"inventory-system/internal/domain"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"net/http"
	"strconv"
)

type SaleApiHandler struct {
	BaseHandler
	saleService service.SaleService
}

func NewSaleApiHandler(base BaseHandler, sService service.SaleService) *SaleApiHandler {
	return &SaleApiHandler{
		BaseHandler: base,
		saleService: sService,
	}
}

func (h *SaleApiHandler) Store(w http.ResponseWriter, r *http.Request) {
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
		PaymentMethod      string            `json:"payment_method"`
		CustomerName       string            `json:"customer_name"`
		DoctorName         string            `json:"doctor_name"`
		PrescriptionNumber string            `json:"prescription_number"`
		Discount           float64           `json:"discount"`
		Items              []domain.SaleItem `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	saleID, err := h.saleService.ProcessSale(session.UserID, data.Items, data.PaymentMethod, data.CustomerName, data.DoctorName, data.PrescriptionNumber, data.Discount)
	if err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"sale_id": saleID,
		"message": "Sale completed successfully",
	})
}

func (h *SaleApiHandler) Void(w http.ResponseWriter, r *http.Request) {
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
		SaleID int    `json:"sale_id"`
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if session.Role == "admin" {
		if err := h.saleService.VoidSale(data.SaleID, session.UserID); err != nil {
			h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	} else {
		// Staff request void
		if err := h.saleService.RequestVoid(data.SaleID, session.UserID, data.Reason); err != nil {
			h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Void processed successfully",
	})
}

func (h *SaleApiHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	saleIDStr := r.URL.Query().Get("id")
	saleID, _ := strconv.Atoi(saleIDStr)

	details, err := h.saleService.GetSaleDetails(saleID)
	if err != nil {
		h.RespondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, details)
}

func (h *SaleApiHandler) GetDetailsHTML(w http.ResponseWriter, r *http.Request) {
	saleIDStr := r.URL.Query().Get("id")
	saleID, _ := strconv.Atoi(saleIDStr)

	details, err := h.saleService.GetSaleDetails(saleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.RenderStandalone(w, r, "reports/history_detail", details)
}
