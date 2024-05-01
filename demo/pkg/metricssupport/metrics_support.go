package metricssupport

import (
	"net/http"
	"strings"
)

type metricsHandler struct {
}

func (h metricsHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}

func MetricsHandler() http.Handler {
	return metricsHandler{}
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, s := range []string{"/styles", "/images"} {
			if strings.HasPrefix(r.URL.Path, s) {
				next.ServeHTTP(w, r)
				return
			}
		}
		// route := mux.CurrentRoute(r)
		// path, _ := route.GetPathTemplate()
		// log.Printf("Collecting metrics for path %v\n", path)
		next.ServeHTTP(w, r)
	})
}
