package handler

import (
	"inventory-system/internal/i18n"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"net/http"
)

type DashboardHandler struct {
	BaseHandler
	ReportService  service.ReportService
	AdminService   service.AdminService
	ProductService service.ProductService
	BatchService   service.BatchService
}

func NewDashboardHandler(base BaseHandler, rService service.ReportService, aService service.AdminService, pService service.ProductService, bService service.BatchService) *DashboardHandler {
	return &DashboardHandler{
		BaseHandler:    base,
		ReportService:  rService,
		AdminService:   aService,
		ProductService: pService,
		BatchService:   bService,
	}
}

func (h *DashboardHandler) Index(w http.ResponseWriter, r *http.Request) {
	lang := h.GetLang(r)
	data := map[string]interface{}{
		"pageTitle":        i18n.T("dashboard", lang),
		"browserTitle":     i18n.T("dashboard", lang) + " - " + "Inventory System",
		"currentPage":      "dashboard",
		"todayRevenue":     0.0,
		"todaySales":       0,
		"revenueDelta":     0.0,
		"salesDelta":       0,
		"lowStockCount":    0,
		"expiringCount":    0,
		"expiredCount":     0,
		"todayProfit":      0.0,
		"recentActivities": []interface{}{},
		"weeklySales": map[string]interface{}{
			"labels": []string{},
			"values": []float64{},
		},
	}

	hubData, err := h.ReportService.GetReportHubData()
	if err == nil {
		for k, v := range hubData {
			data[k] = v
		}

		// Calculate revenue delta percentage
		if todayRev, ok := hubData["todayRevenue"].(float64); ok {
			if delta, ok := hubData["revenueDelta"].(float64); ok {
				yesterdayRev := todayRev - delta
				if yesterdayRev > 0 {
					data["revenueDeltaPct"] = (delta / yesterdayRev) * 100
				} else if yesterdayRev == 0 && todayRev > 0 {
					data["revenueDeltaPct"] = 100.0
				} else {
					data["revenueDeltaPct"] = 0.0
				}
			}
		}
	}

	health, err := h.AdminService.GetSystemHealth()
	if err == nil {
		data["SystemHealth"] = health
	}

	recentProducts, _ := h.ProductService.GetRecentProducts(5)
	data["recentProducts"] = recentProducts

	expiringCount, _ := h.BatchService.GetExpiringCount(180) // 180 days warning matching mockup
	data["expiringCount"] = expiringCount

	h.Render(w, r, "dashboard/index", data)
}

func (h *DashboardHandler) Health(w http.ResponseWriter, r *http.Request) {
	health, err := h.AdminService.GetSystemHealth()
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		data := map[string]interface{}{
			"SystemHealth": health,
			"Lang":         "en",
		}
		session := middleware.GetSession(r)
		if session != nil {
			data["Lang"] = session.Lang
		}
		h.RenderStandalone(w, r, "layouts/health_widget", data)
		return
	}

	h.RespondJSON(w, http.StatusOK, health)
}

func (h *DashboardHandler) GetHubData(w http.ResponseWriter, r *http.Request) {
	hubData, err := h.ReportService.GetReportHubData()
	if err != nil {
		h.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.RespondJSON(w, http.StatusOK, hubData)
}
