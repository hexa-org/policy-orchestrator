package azuresupport

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

type AzureSupport struct {
	session *sessions.CookieStore
}

func NewAzureSupport(session *sessions.CookieStore) *AzureSupport {
	log.Println("Enabling azure authn/z middleware.")
	return &AzureSupport{session}
}

func (g *AzureSupport) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if name := r.Header["X-Ms-Client-Principal-Name"]; name != nil && len(name) > 0 {
			log.Println("Found azure authenticated user.")
			session, _ := g.session.Get(r, "session")
			session.Values["principal"] = name
			session.Values["logout"] = "/.auth/logout"
			err := session.Save(r, w)
			if err == nil {
				log.Println("Saved authenticated user name in session.")
			}
		}
		next.ServeHTTP(w, r)
	})
}
