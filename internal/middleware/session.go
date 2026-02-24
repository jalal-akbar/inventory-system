package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func init() {

	if os.Getenv("SESSION_KEY") == "" {
		Store = sessions.NewCookieStore([]byte("very-secret-key-change-it-in-prod"))
	}

	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0,     // 0 session will be deleted when the browser is closed
		HttpOnly: true,  // cookie cannot be read by script JS
		Secure:   false, // Set to true later if using HTTPS
		SameSite: http.SameSiteLaxMode,
	}
}

type SessionData struct {
	UserID   int
	Username string
	Role     string
	Lang     string
}

func GetSession(r *http.Request) *SessionData {
	session, _ := Store.Get(r, "inventory-session")
	uid, ok1 := session.Values["user_id"].(int)
	uname, ok2 := session.Values["username"].(string)
	role, ok3 := session.Values["role"].(string)
	lang, ok4 := session.Values["lang"].(string)

	if !ok1 || !ok2 || !ok3 {
		return nil
	}
	if !ok4 {
		lang = "en"
	}

	return &SessionData{
		UserID:   uid,
		Username: uname,
		Role:     role,
		Lang:     lang,
	}
}

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := GetSession(r)
		if data != nil {
			ctx := context.WithValue(r.Context(), "user", data)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
