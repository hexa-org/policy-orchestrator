package websupport

import (
	"context"
	"crypto/tls"

	"net"
	"net/http"
	"os"
	"strings"

	log "golang.org/x/exp/slog"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/metricssupport"
)

const EnvTlsEnabled string = "HEXA_TLS_ENABLED"

type Options struct {
	HealthChecks []healthsupport.HealthCheck
}

type Path struct {
	URI     string
	Methods []string
}

func Paths(router *mux.Router) []Path {
	var paths []Path
	_ = router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		uri, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		paths = append(paths, Path{uri, methods})
		return nil
	})
	return paths
}

func Create(addr string, handlers func(x *mux.Router), options Options) *http.Server {
	checks := options.HealthChecks
	if checks == nil || len(checks) == 0 {
		checks = append(checks, &healthsupport.NoopCheck{})
	}

	router := mux.NewRouter()
	router.Use(metricssupport.MetricsMiddleware)
	router.HandleFunc("/health",
		func(w http.ResponseWriter, r *http.Request) {
			healthsupport.HealthHandlerFunctionWithChecks(w, r, checks)
		},
	).Methods("GET")
	router.Path("/metrics").Handler(metricssupport.MetricsHandler())
	router.StrictSlash(true)
	handlers(router)
	server := http.Server{
		Addr:    addr,
		Handler: router,
	}
	for _, p := range Paths(router) {
		log.Debug("Registered route", p.Methods, p.URI)
	}
	return &server
}

func Start(server *http.Server, l net.Listener) {
	if server.TLSConfig != nil {
		log.Info("Starting the server with tls support", "address", server.Addr)
		err := server.ServeTLS(l, "", "")
		if err != nil {
			log.Error("error starting the server:", err.Error())
			return
		}
	}

	log.Info("Starting the server", "address", server.Addr)
	err := server.Serve(l)
	if err != nil {
		log.Error("error starting the server:", err.Error())
		return
	}
}

func Stop(server *http.Server) {
	log.Info("Stopping the server.")
	_ = server.Shutdown(context.Background())
}

func IsTlsEnabled() bool {
	mode := os.Getenv(EnvTlsEnabled)
	if mode != "" && (strings.EqualFold(mode[0:1], "y") || strings.EqualFold(mode[0:1], "t")) {
		return true
	}
	return false
}

func WithTransportLayerSecurity(certFile, keyFile string, app *http.Server) {
	if IsTlsEnabled() {
		cert, certErr := os.ReadFile(certFile)
		if certErr != nil {
			panic(certErr.Error())
		}
		key, keyErr := os.ReadFile(keyFile)
		if keyErr != nil {
			panic(certErr.Error())
		}
		pair, pairErr := tls.X509KeyPair(cert, key)
		if pairErr != nil {
			panic(pairErr.Error())
		}
		app.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{pair},
		}
		return
	}
}
