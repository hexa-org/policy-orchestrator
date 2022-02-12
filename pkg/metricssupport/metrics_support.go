package metricssupport

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var httpDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{Name: "any_request_duration_seconds"}, []string{"path"})

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		log.Println(r.Header)
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}
