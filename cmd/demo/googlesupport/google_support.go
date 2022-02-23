package googlesupport

import (
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

type GoogleSupport struct {
	session *sessions.CookieStore
}

func NewGoogleSupport(session *sessions.CookieStore) *GoogleSupport {
	log.Println("Enabling google iap middleware.")
	return &GoogleSupport{session}
}

func (g *GoogleSupport) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if email := r.Header["X-Goog-Authenticated-User-Email"]; email != nil && len(email) > 0 {
			log.Println("Found google authenticated user email.")
			session, _ := g.session.Get(r, "session")
			session.Values["principal"] = email
			session.Values["logout"] = "?gcp-iap-mode=CLEAR_LOGIN_COOKIE"
			err := session.Save(r, w)
			if err == nil {
				log.Println("Saved authenticated user email in session.")
			}
		}
		next.ServeHTTP(w, r)
	})
}
