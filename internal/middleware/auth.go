package middleware

import (
	"net/http"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		if session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		if session == nil || session.Role != "admin" {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
