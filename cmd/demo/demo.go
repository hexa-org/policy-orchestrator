package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/opasupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func App(client HTTPClient, opaUrl string, addr string, resourcesDirectory string) *http.Server {
	opaSupport, err := opasupport.NewOpaSupport(client, opaUrl)
	if err != nil {
		log.Fatalln(err)
	}
	server := websupport.Create(addr, loadHandlers(opaSupport), websupport.Options{ResourceDirectory: resourcesDirectory})
	return server
}

func dashboard(req http.ResponseWriter, _ *http.Request) {
	_ = websupport.ModelAndView(req, "dashboard", websupport.Model{Map: map[string]interface{}{}})
}

func accounting(writer http.ResponseWriter, _ *http.Request) {
	_ = websupport.ModelAndView(writer, "accounting", websupport.Model{Map: map[string]interface{}{}})
}

func sales(writer http.ResponseWriter, _ *http.Request) {
	_ = websupport.ModelAndView(writer, "sales", websupport.Model{Map: map[string]interface{}{}})
}

func humanresources(writer http.ResponseWriter, _ *http.Request) {
	_ = websupport.ModelAndView(writer, "humanresources", websupport.Model{Map: map[string]interface{}{}})
}

func unauthorized(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusUnauthorized)
	_ = websupport.ModelAndView(writer, "unauthorized", websupport.Model{Map: map[string]interface{}{}})
}

func download(writer http.ResponseWriter, _ *http.Request) {
	_, file, _, _ := runtime.Caller(0)
	opasupport.Compress(writer, filepath.Join(file, "../resources/bundles/bundle"))
}

func loadHandlers(opa *opasupport.OpaSupport) func(router *mux.Router) {
	return func(router *mux.Router) {
		router.HandleFunc("/", opasupport.OpaMiddleware(opa, dashboard, unauthorized)).Methods("GET")
		router.HandleFunc("/sales", opasupport.OpaMiddleware(opa, sales, unauthorized)).Methods("GET")
		router.HandleFunc("/accounting", opasupport.OpaMiddleware(opa, accounting, unauthorized)).Methods("GET")
		router.HandleFunc("/humanresources", opasupport.OpaMiddleware(opa, humanresources, unauthorized)).Methods("GET")
		router.HandleFunc("/bundles/bundle.tar.gz", download).Methods("GET")

		fileServer := http.FileServer(http.Dir("cmd/demo/resources/static"))
		router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	}
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Printf("Found server address %v", addr)

	opaUrl := "http://0.0.0.0:8887/v1/data/authz/allow"
	if found := os.Getenv("OPA_SERVER_URL"); found != "" {
		opaUrl = found
	}
	log.Printf("Found open policy agenet server address %v", opaUrl)

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../cmd/demo/resources")
	listener, _ := net.Listen("tcp", addr)
	return App(&http.Client{}, opaUrl, listener.Addr().String(), resourcesDirectory), listener
}

func main() {
	websupport.Start(newApp("0.0.0.0:8886"))
}
