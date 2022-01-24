package websupport

import (
	"net/http"
)

func SecurityContextViewSupport(w http.ResponseWriter, r *http.Request, view string, model Model) (http.ResponseWriter, string, Model) {
	if email := r.Header["X-Goog-Authenticated-User-Email"]; email != nil && len(email) > 0 {
		model.Map["email"] = email
	}
	return w, view, model
}
