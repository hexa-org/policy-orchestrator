package googlesupport

import (
	"net/http"
	"github.com/gorilla/sessions"
)

type GoogleSupport struct {
	session *sessions.CookieStore
}

func NewGoogleSupport(session *sessions.CookieStore) *GoogleSupport {
	return &GoogleSupport{session}
}

func (g *GoogleSupport) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if email := r.Header["X-Goog-Authenticated-User-Email"]; email != nil && len(email) > 0 {
			session, _ := g.session.Get(r, "session")
			session.Values["principal"] =  email
			_ = session.Save(r, w)
		}
		next.ServeHTTP(w, r)
	})
}
