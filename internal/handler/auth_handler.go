package handler

import (
	"inventory-system/internal/domain"
	"inventory-system/internal/i18n"
	"inventory-system/internal/middleware"
	"inventory-system/internal/service"
	"log"
	"net/http"
)

type AuthHandler struct {
	BaseHandler
	AuthService service.AuthService
}

func NewAuthHandler(base BaseHandler, authService service.AuthService) *AuthHandler {
	return &AuthHandler{BaseHandler: base, AuthService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if middleware.GetSession(r) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	errorMsg := ""
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		tz := r.FormValue("detected_timezone")

		user, err := h.AuthService.Login(username, password)
		if err != nil {
			lang := "en" // Default for login
			switch err {
			case domain.ErrInvalidCredentials:
				errorMsg = i18n.T("err_invalid_credentials", lang)
			case domain.ErrAccountDisabled:
				errorMsg = i18n.T("err_account_disabled", lang)
			default:
				log.Printf("Login error: %v", err)
				errorMsg = i18n.T("err_unexpected_error", lang)
			}
		} else {
			session, _ := middleware.Store.Get(r, "inventory-session")
			session.Values["user_id"] = user.ID
			session.Values["username"] = user.Username
			session.Values["role"] = user.Role
			session.Values["lang"] = user.Language
			if tz != "" {
				session.Values["timezone"] = tz
			}
			session.Save(r, w)

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	h.RenderAuth(w, r, "auth/login", map[string]interface{}{"Error": errorMsg})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.Store.Get(r, "inventory-session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
