package handler

import (
	"fmt"
	"inventory-system/internal/i18n"
	"inventory-system/internal/service"
	"net/http"
	"time"
)

type ReportHandler struct {
	BaseHandler
	ReportService service.ReportService
}

func NewReportHandler(base BaseHandler, rService service.ReportService) *ReportHandler {
	return &ReportHandler{
		BaseHandler:   base,
		ReportService: rService,
	}
}

func (h *ReportHandler) Hub(w http.ResponseWriter, r *http.Request) {
	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":    i18n.T("reports", lang),
		"browserTitle": i18n.T("reports", lang) + " - " + "Inventory System",
		"currentPage":  "reports",
	}

	hubData, err := h.ReportService.GetReportHubData()
	if err == nil {
		for k, v := range hubData {
			data[k] = v
		}
	}

	h.Render(w, r, "reports/hub", data)
}

func (h *ReportHandler) Financial(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	category := r.URL.Query().Get("category")

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	reportData, err := h.ReportService.GetFinancialReportData(startDate, endDate, category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":    i18n.T("financial", lang),
		"browserTitle": i18n.T("financial", lang) + " - " + "Inventory System",
		"currentPage":  "reports",
		"startDate":    startDate,
		"endDate":      endDate,
		"category":     category,
	}
	for k, v := range reportData {
		data[k] = v
	}

	h.Render(w, r, "reports/index", data)
}

func (h *ReportHandler) History(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	history, err := h.ReportService.GetHistory(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":    i18n.T("sales_history", lang),
		"browserTitle": i18n.T("sales_history", lang) + " - " + "Inventory System",
		"currentPage":  "sales_history",
		"startDate":    startDate,
		"endDate":      endDate,
		"sales":        history,
	}

	h.Render(w, r, "reports/history", data)
}

func (h *ReportHandler) Psychotropic(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	report, err := h.ReportService.GetPsychotropicReport(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":    i18n.T("psychotropic", lang),
		"browserTitle": i18n.T("psychotropic", lang) + " - " + "Inventory System",
		"currentPage":  "reports",
		"startDate":    startDate,
		"endDate":      endDate,
		"report":       report,
	}

	h.Render(w, r, "reports/psychotropic", data)
}

func (h *ReportHandler) StockMutation(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	mutation, err := h.ReportService.GetStockMutation(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":    i18n.T("stock_mutation", lang),
		"browserTitle": i18n.T("stock_mutation", lang) + " - " + "Inventory System",
		"currentPage":  "reports",
		"startDate":    startDate,
		"endDate":      endDate,
		"mutation":     mutation,
	}

	h.Render(w, r, "reports/stock_mutation", data)
}

func (h *ReportHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	category := r.URL.Query().Get("category")

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=sales_report_%s.csv", time.Now().Format("20060102")))

	if err := h.ReportService.ExportRecentSalesCSV(w, startDate, endDate, category); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ReportHandler) ExportStockMutationCSV(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=stock_mutation_%s.csv", time.Now().Format("20060102")))

	if err := h.ReportService.ExportStockMutationCSV(w, startDate, endDate); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
