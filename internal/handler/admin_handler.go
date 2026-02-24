package handler

import (
	"encoding/json"
	"inventory-system/internal/i18n"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"net/http"
	"strconv"
	"time"
)

type AdminHandler struct {
	BaseHandler
	AdminService service.AdminService
}

func NewAdminHandler(base BaseHandler, aService service.AdminService) *AdminHandler {
	return &AdminHandler{
		BaseHandler:  base,
		AdminService: aService,
	}
}

func (h *AdminHandler) Approvals(w http.ResponseWriter, r *http.Request) {
	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":    i18n.T("admin_approval_dashboard", lang),
		"browserTitle": i18n.T("admin_approval_dashboard", lang) + " - " + "Inventory System",
		"currentPage":  "approvals",
	}
	h.Render(w, r, "admin/approvals", data)
}

func (h *AdminHandler) GetPendingItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.AdminService.GetPendingItems()
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (h *AdminHandler) ApproveItem(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	session := middleware.GetSession(r)
	if err := h.AdminService.ApproveItem(body.ID, body.Type, session.UserID); err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) RejectItem(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	session := middleware.GetSession(r)
	if err := h.AdminService.RejectItem(body.ID, body.Type, session.UserID); err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) ApproveGroup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProductID int `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	session := middleware.GetSession(r)
	if err := h.AdminService.ApproveGroup(body.ProductID, session.UserID); err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) ApproveAll(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)
	if err := h.AdminService.ApproveAll(session.UserID); err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AdminHandler) Backup(w http.ResponseWriter, r *http.Request) {
	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":    i18n.T("system_backup", lang),
		"browserTitle": i18n.T("system_backup", lang) + " - " + "Inventory System",
		"currentPage":  "backup",
	}
	h.Render(w, r, "admin/backup", data)
}

func (h *AdminHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	path, err := h.AdminService.CreateBackup()
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.FormatInt(time.Now().Unix(), 10)+"_backup.sql")
	http.ServeFile(w, r, path)
}
