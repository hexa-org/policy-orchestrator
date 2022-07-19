package websupport

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/healthsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/metricssupport"
	"log"
	"net"
	"net/http"
)

type Options struct {
	ResourceDirectory string
	HealthChecks      []healthsupport.HealthCheck
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
	resourcesDirectory = options.ResourceDirectory
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
		log.Println("Registered route", p.Methods, p.URI)
	}
	return &server
}

func Start(server *http.Server, l net.Listener) {
	log.Println("Starting the server.", server.Addr)
	err := server.Serve(l)
	if err != nil {
		return
	}
}

func StartWithTLS(server *http.Server, l net.Listener) {
	if server.TLSConfig == nil {
		panic("tls not configured.")
	}
	log.Println("Starting the server with tls support.", server.Addr)
	err := server.ServeTLS(l, "", "")
	if err != nil {
		return
	}
}

func Stop(server *http.Server) {
	log.Printf("Stopping the server.")
	_ = server.Shutdown(context.Background())
}
