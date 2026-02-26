package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"inventory-system/internal/domain"
	"inventory-system/internal/i18n"
	"inventory-system/internal/middleware"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type BaseHandler struct {
	SettingService service.SettingService
	PendingRepo    repository.PendingCountsRepository
}

func (h *BaseHandler) GetLang(r *http.Request) string {
	lang := "en"
	session := middleware.GetSession(r)
	if session != nil {
		lang = session.Lang
	}
	return lang
}

func (h *BaseHandler) getFuncMap(tz string) template.FuncMap {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	return template.FuncMap{
		"t": i18n.T,
		"formatRupiah": func(amount interface{}) string {
			f, ok := amount.(float64)
			if !ok {
				// Try int
				if i, ok := amount.(int); ok {
					f = float64(i)
				} else if i64, ok := amount.(int64); ok {
					f = float64(i64)
				}
			}
			return fmt.Sprintf("Rp %s", formatNumber(f))
		},
		"formatNumber": func(amount interface{}) string {
			f, ok := amount.(float64)
			if !ok {
				if i, ok := amount.(int); ok {
					f = float64(i)
				} else if i64, ok := amount.(int64); ok {
					f = float64(i64)
				}
			}
			return formatNumber(f)
		},
		"formatDateTime": func(t interface{}) string {
			var tm time.Time
			switch v := t.(type) {
			case time.Time:
				tm = v
			case string:
				if v == "now" {
					tm = time.Now()
				} else {
					tm, _ = time.Parse(time.RFC3339, v)
				}
			default:
				return ""
			}
			return tm.In(loc).Format("02 Jan 2006, 15:04:05")
		},
		"formatShortDate": func(t interface{}) string {
			var tm time.Time
			switch v := t.(type) {
			case time.Time:
				tm = v
			case string:
				if v == "today" {
					tm = time.Now()
				} else {
					tm, _ = time.Parse("2006-01-02", v)
				}
			default:
				return ""
			}
			return tm.In(loc).Format("02-01-2006")
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": func(s string) string {
			if s == "" {
				return ""
			}
			return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
		},
		"firstChar": func(s string) string {
			if s == "" {
				return "?"
			}
			return strings.ToUpper(s[:1])
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b interface{}) float64 {
			var fa, fb float64
			switch v := a.(type) {
			case int:
				fa = float64(v)
			case int64:
				fa = float64(v)
			case float64:
				fa = v
			}
			switch v := b.(type) {
			case int:
				fb = float64(v)
			case int64:
				fb = float64(v)
			case float64:
				fb = v
			}
			return fa * fb
		},
		"div": func(a, b interface{}) float64 {
			var fa, fb float64
			switch v := a.(type) {
			case int:
				fa = float64(v)
			case int64:
				fa = float64(v)
			case float64:
				fa = v
			}
			switch v := b.(type) {
			case int:
				fb = float64(v)
			case int64:
				fb = float64(v)
			case float64:
				fb = v
			}
			if fb == 0 {
				return 0
			}
			return fa / fb
		},
		"abs": func(a interface{}) float64 {
			var f float64
			switch v := a.(type) {
			case int:
				f = float64(v)
			case int64:
				f = float64(v)
			case float64:
				f = v
			}
			if f < 0 {
				return -f
			}
			return f
		},
		"json": func(v interface{}) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("[]")
			}
			return template.JS(b)
		},
		"formatPercentage": func(v float64) string {
			return fmt.Sprintf("%.1f%%", v)
		},
		"formatPcsToUnit": func(totalPcs int, unitName string, itemsPerUnit int) string {
			if totalPcs <= 0 {
				return fmt.Sprintf("0 %s", unitName)
			}
			if itemsPerUnit <= 1 {
				// No conversion needed
				return fmt.Sprintf("%d %s", totalPcs, unitName)
			}

			units := totalPcs / itemsPerUnit
			remainingPcs := totalPcs % itemsPerUnit

			if units > 0 && remainingPcs > 0 {
				return fmt.Sprintf("%d %s %d PCS", units, unitName, remainingPcs)
			} else if units > 0 {
				return fmt.Sprintf("%d %s", units, unitName)
			}
			return fmt.Sprintf("%d PCS", remainingPcs)
		},
	}
}

func (h *BaseHandler) Render(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	// Global Data
	settings, err := h.SettingService.GetSettings()
	if err != nil {
		log.Printf("Error getting settings: %v", err)
	}
	if settings == nil {
		settings = &domain.Setting{BusinessName: "Inventory System"}
	}
	data["GlobalSettings"] = settings

	pending, _ := h.PendingRepo.GetCounts()
	if pending == nil {
		pending = &repository.PendingCounts{} // Prevent template panic
	}
	data["PendingCounts"] = pending

	session := middleware.GetSession(r)
	data["CurrentUser"] = session
	data["CSRFToken"] = middleware.GetCSRFToken(r)

	// i18n helper
	data["Lang"] = "en"
	if session != nil {
		data["Lang"] = session.Lang
	}

	// Templates
	layoutDir := "web/templates/layouts"
	pagePath := filepath.Join("web/templates", page+".html")
	pageDir := filepath.Dir(pagePath)

	tz := "UTC"
	if settings != nil && settings.Timezone != "" {
		tz = settings.Timezone
	}
	funcMap := h.getFuncMap(tz)

	files := []string{
		filepath.Join(layoutDir, "base.html"),
		filepath.Join(layoutDir, "topbar.html"),
		filepath.Join(layoutDir, "sidebar.html"),
		filepath.Join(layoutDir, "footer.html"),
		filepath.Join(layoutDir, "health_widget.html"),
		pagePath,
	}

	// Also include any other partial .html files in the same directory (must start with _)
	partials, _ := filepath.Glob(filepath.Join(pageDir, "_*.html"))
	for _, p := range partials {
		exists := false
		for _, f := range files {
			if f == p {
				exists = true
				break
			}
		}
		if !exists {
			files = append(files, p)
		}
	}

	tmpl, err := template.New(filepath.Base(pagePath)).Funcs(funcMap).ParseFiles(files...)

	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *BaseHandler) RenderAuth(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	settings, err := h.SettingService.GetSettings()
	if err != nil {
		log.Printf("Error getting auth settings: %v", err)
	}
	if settings == nil {
		settings = &domain.Setting{BusinessName: "Inventory System"}
	}
	data["GlobalSettings"] = settings
	data["CSRFToken"] = middleware.GetCSRFToken(r)

	// Default to en for login if no session (which is always for login)
	data["Lang"] = "en"

	pagePath := filepath.Join("web/templates", page+".html")
	tz := "UTC"
	if settings != nil && settings.Timezone != "" {
		tz = settings.Timezone
	}
	tmpl, err := template.New(filepath.Base(pagePath)).Funcs(h.getFuncMap(tz)).ParseFiles(pagePath)
	if err != nil {
		log.Printf("Auth template parse error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Auth template execution error: %v", err)
	}
}

func (h *BaseHandler) RenderStandalone(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	// Global Data
	settings, err := h.SettingService.GetSettings()
	if err != nil {
		log.Printf("Error getting settings: %v", err)
	}
	if settings == nil {
		settings = &domain.Setting{BusinessName: "Inventory System"}
	}
	data["GlobalSettings"] = settings

	pending, _ := h.PendingRepo.GetCounts()
	if pending == nil {
		pending = &repository.PendingCounts{} // Prevent template panic
	}
	data["PendingCounts"] = pending

	session := middleware.GetSession(r)
	data["CurrentUser"] = session
	data["CSRFToken"] = middleware.GetCSRFToken(r)

	data["Lang"] = "en"
	if session != nil {
		data["Lang"] = session.Lang
	}

	// Templates
	pagePath := filepath.Join("web/templates", page+".html")
	pageDir := filepath.Dir(pagePath)

	tz := "UTC"
	if settings != nil && settings.Timezone != "" {
		tz = settings.Timezone
	}
	funcMap := h.getFuncMap(tz)

	files := []string{pagePath}
	// Also include any other partial .html files in the same directory (must start with _)
	partials, _ := filepath.Glob(filepath.Join(pageDir, "_*.html"))
	for _, p := range partials {
		exists := false
		for _, f := range files {
			if f == p {
				exists = true
				break
			}
		}
		if !exists {
			files = append(files, p)
		}
	}

	tmpl, err := template.New(filepath.Base(pagePath)).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *BaseHandler) RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func formatNumber(f float64) string {
	s := fmt.Sprintf("%.0f", f)
	if f == 0 {
		return "0"
	}
	var res []string
	for i := len(s); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		res = append([]string{s[start:i]}, res...)
	}
	return strings.Join(res, ".")
}
