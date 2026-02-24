package handler

import (
	"inventory-system/internal/i18n"
	"inventory-system/internal/middleware"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"net/http"
	"strconv"
)

type UserHandler struct {
	BaseHandler
	UserService    service.UserService
	SettingService service.SettingService
	LogRepo        repository.ActivityLogRepository
}

func NewUserHandler(base BaseHandler, userService service.UserService, settingService service.SettingService, logRepo repository.ActivityLogRepository) *UserHandler {
	return &UserHandler{BaseHandler: base, UserService: userService, SettingService: settingService, LogRepo: logRepo}
}

func (h *UserHandler) Index(w http.ResponseWriter, r *http.Request) {
	msg := ""
	err_msg := ""

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		lang := "en"
		session := middleware.GetSession(r)
		if session != nil {
			lang = session.Lang
		}

		switch action {
		case "add":
			err := h.UserService.Create(r.FormValue("username"), r.FormValue("password"), r.FormValue("role"))
			if err != nil {
				err_msg = err.Error()
			} else {
				msg = i18n.T("user_added_success", lang)
			}
		case "edit":
			id, _ := strconv.Atoi(r.FormValue("user_id"))
			input := service.UpdateUserInput{
				Username: r.FormValue("username"),
				Role:     r.FormValue("role"),
				Status:   r.FormValue("status"),
				Password: r.FormValue("password"),
			}
			err := h.UserService.Update(id, input)
			if err != nil {
				err_msg = err.Error()
			} else {
				msg = i18n.T("user_updated_success", lang)
			}
		case "delete":
			id, _ := strconv.Atoi(r.FormValue("user_id"))
			h.UserService.SetStatus(id, "inactive")
			msg = i18n.T("user_inactivated", lang)
		case "activate":
			id, _ := strconv.Atoi(r.FormValue("user_id"))
			h.UserService.SetStatus(id, "active")
			msg = i18n.T("user_activated", lang)
		}
	}

	users, _ := h.UserService.GetAll()
	lang := "en"
	session := middleware.GetSession(r)
	if session != nil {
		lang = session.Lang
	}

	h.Render(w, r, "users/index", map[string]interface{}{
		"Users":       users,
		"Message":     msg,
		"Error":       err_msg,
		"PageTitle":   i18n.T("manage_users", lang),
		"CurrentPage": "users",
	})
}

func (h *UserHandler) Settings(w http.ResponseWriter, r *http.Request) {
	lang := "en"
	session := middleware.GetSession(r)
	if session != nil {
		lang = session.Lang
	}

	// Fetch all data for the consolidated settings page
	users, _ := h.UserService.GetAll()
	globalSettings, _ := h.SettingService.GetSettings()
	logs, _ := h.LogRepo.Search("", "", "desc") // Default to newest first

	h.Render(w, r, "admin/settings", map[string]interface{}{
		"PageTitle":      i18n.T("settings", lang),
		"CurrentPage":    "settings",
		"ActiveTab":      r.URL.Query().Get("tab"),
		"Users":          users,
		"GlobalSettings": globalSettings,
		"Logs":           logs,
		"Message":        r.URL.Query().Get("message"),
		"Error":          r.URL.Query().Get("error"),
	})
}

func (h *UserHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/settings/business", http.StatusSeeOther)
		return
	}

	session := middleware.GetSession(r)
	if session == nil || session.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	businessName := r.FormValue("business_name")
	address := r.FormValue("address")
	phone := r.FormValue("phone")
	currencySymbol := r.FormValue("currency_symbol")
	timezone := r.FormValue("timezone")

	err := h.SettingService.UpdateGlobalSettings(session.UserID, businessName, address, phone, currencySymbol, timezone)
	if err != nil {
		// Just log and redirect for now, could add flash message
		http.Redirect(w, r, "/settings?tab=business&error="+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/settings?tab=business&message=Settings+updated+successfully", http.StatusSeeOther)
}

func (h *UserHandler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := middleware.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	newUsername := r.FormValue("username")
	if newUsername == "" {
		http.Redirect(w, r, "/account-settings?error=Username+cannot+be+empty", http.StatusSeeOther)
		return
	}

	// Update username logic
	// We'll use h.UserService.Update which takes UpdateUserInput
	err := h.UserService.Update(session.UserID, service.UpdateUserInput{
		Username: newUsername,
	})

	if err != nil {
		http.Redirect(w, r, "/settings?tab=account&error="+err.Error(), http.StatusSeeOther)
		return
	}

	// Update session username if possible (depends on session implementation)
	// For now, redirect with success message
	http.Redirect(w, r, "/settings?tab=account&message=Username+updated+successfully", http.StatusSeeOther)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := middleware.GetSession(r)
	if session == nil || session.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	err := h.UserService.Create(username, password, role)
	if err != nil {
		http.Redirect(w, r, "/settings?tab=users&error="+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/settings?tab=users&message=User+created+successfully", http.StatusSeeOther)
}

func (h *UserHandler) ToggleUserStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := middleware.GetSession(r)
	if session == nil || session.Role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := strconv.Atoi(r.FormValue("user_id"))

	// Get current user to check status
	users, _ := h.UserService.GetAll()
	var currentUserStatus string
	for _, u := range users {
		if u.ID == userID {
			currentUserStatus = u.Status
			break
		}
	}

	newStatus := "active"
	if currentUserStatus == "active" {
		newStatus = "inactive"
	}

	h.UserService.SetStatus(userID, newStatus)

	http.Redirect(w, r, "/settings?tab=users&message=User+status+updated", http.StatusSeeOther)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := middleware.GetSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	oldPwd := r.FormValue("current_password")
	newPwd := r.FormValue("new_password")
	cfmPwd := r.FormValue("confirm_password")

	if newPwd != cfmPwd {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "New passwords do not match"})
		return
	}

	err := h.UserService.ChangePassword(session.UserID, oldPwd, newPwd)
	if err != nil {
		h.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

func (h *UserHandler) SetLanguage(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = r.FormValue("lang")
	}

	if lang != "en" && lang != "id" {
		lang = "en"
	}

	session := middleware.GetSession(r)
	if session != nil {
		h.UserService.UpdateLanguage(session.UserID, lang)
		// Update session cookie
		sess, _ := middleware.Store.Get(r, "inventory-session")
		sess.Values["lang"] = lang
		sess.Save(r, w)
	}

	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/dashboard"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}
