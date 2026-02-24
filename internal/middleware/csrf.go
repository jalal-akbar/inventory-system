package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func CSRFProtect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "inventory-session")

		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			token := r.FormValue("csrf_token")
			if token == "" {
				token = r.Header.Get("X-CSRF-Token")
			}

			storedToken, ok := session.Values["csrf_token"].(string)
			if !ok || token == "" || token != storedToken {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)
				return
			}
		}

		// Ensure token exists for GET requests
		if _, ok := session.Values["csrf_token"].(string); !ok {
			token := make([]byte, 32)
			rand.Read(token)
			session.Values["csrf_token"] = hex.EncodeToString(token)
			session.Save(r, w)
		}

		next.ServeHTTP(w, r)
	})
}

func GetCSRFToken(r *http.Request) string {
	session, _ := Store.Get(r, "inventory-session")
	token, _ := session.Values["csrf_token"].(string)
	return token
}
