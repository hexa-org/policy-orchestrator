package websupport

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
)

type Options struct {
	ResourceDirectory string
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

type Health struct {
	Status string `json:"status"`
}

func health(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(&Health{"pass"})
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func Create(addr string, handlers func(x *mux.Router), options Options) *http.Server {
	resourcesDirectory = options.ResourceDirectory

	router := mux.NewRouter()
	router.HandleFunc("/health", health).Methods("GET")
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

func WaitForHealthy(server *http.Server) {
	var isLive bool
	for !isLive {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", server.Addr))
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Println("Server is healthy.", server.Addr)
			isLive = true
		}
	}
}

func Stop(server *http.Server) {
	log.Printf("Stopping the server.")
	_ = server.Shutdown(context.Background())
}
