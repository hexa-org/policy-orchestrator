package decisionsupport

import (
	"log"
	"net/http"
	"strings"
)

type DecisionSupport struct {
	Provider     DecisionProvider
	Unauthorized http.HandlerFunc
	Skip         []string
}

func (d *DecisionSupport) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, s := range d.Skip {
			if strings.HasPrefix(r.RequestURI, s) {
				next.ServeHTTP(w, r)
				return
			}
		}
		log.Println("Checking authorization.")

		input, inputErr := d.Provider.BuildInput(r)
		if inputErr != nil {
			d.Unauthorized(w, r)
			return
		}

		allow, err := d.Provider.Allow(input)
		if !allow || err != nil {
			d.Unauthorized(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
