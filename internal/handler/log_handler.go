package handler

import (
	"inventory-system/internal/i18n"
	"inventory-system/internal/repository"
	"net/http"
)

type LogHandler struct {
	BaseHandler
	LogRepo repository.ActivityLogRepository
}

func NewLogHandler(base BaseHandler, lRepo repository.ActivityLogRepository) *LogHandler {
	return &LogHandler{
		BaseHandler: base,
		LogRepo:     lRepo,
	}
}

func (h *LogHandler) Index(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	sort := r.URL.Query().Get("sort")

	logs, err := h.LogRepo.Search(startDate, endDate, sort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":    i18n.T("activity_log", lang),
		"browserTitle": i18n.T("activity_log", lang) + " - " + "Inventory System",
		"currentPage":  "activity_log",
		"logs":         logs,
		"filters": map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
			"sort":       sort,
		},
	}
	h.Render(w, r, "admin/logs", data)
}
