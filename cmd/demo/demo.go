package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"hexa/cmd/demo/support"
	"hexa/pkg/opa_support"
	"hexa/pkg/web_support"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func App(client HTTPClient, opaUrl string, addr string, resourcesDirectory string) *http.Server {
	opaSupport, err := opa_support.NewOpaSupport(client, opaUrl)
	if err != nil {
		log.Fatalln(err)
	}
	server := web_support.Create(addr, loadHandlers(opaSupport), web_support.Options{ResourceDirectory: resourcesDirectory})
	return server
}

func dashboard(req http.ResponseWriter, _ *http.Request) {
	_ = web_support.ModelAndView(req, "dashboard", web_support.Model{Map: map[string]interface{}{}})
}

func accounting(writer http.ResponseWriter, _ *http.Request) {
	_ = web_support.ModelAndView(writer, "accounting", web_support.Model{Map: map[string]interface{}{}})
}

func sales(writer http.ResponseWriter, _ *http.Request) {
	_ = web_support.ModelAndView(writer, "sales", web_support.Model{Map: map[string]interface{}{}})
}

func humanresources(writer http.ResponseWriter, _ *http.Request) {
	_ = web_support.ModelAndView(writer, "humanresources", web_support.Model{Map: map[string]interface{}{}})
}

func unauthorized(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusUnauthorized)
	_ = web_support.ModelAndView(writer, "unauthorized", web_support.Model{Map: map[string]interface{}{}})
}

func download(writer http.ResponseWriter, _ *http.Request) {
	_, file, _, _ := runtime.Caller(0)
	support.Compress(writer, filepath.Join(file, "../resources/bundles/bundle"))
}

func loadHandlers(opa *opa_support.OpaSupport) func(router *mux.Router) {
	return func(router *mux.Router) {
		router.HandleFunc("/", opa_support.OpaMiddleware(opa, dashboard, unauthorized)).Methods("GET")
		router.HandleFunc("/sales", opa_support.OpaMiddleware(opa, sales, unauthorized)).Methods("GET")
		router.HandleFunc("/accounting", opa_support.OpaMiddleware(opa, accounting, unauthorized)).Methods("GET")
		router.HandleFunc("/humanresources", opa_support.OpaMiddleware(opa, humanresources, unauthorized)).Methods("GET")
		router.HandleFunc("/bundles/bundle.tar.gz", download).Methods("GET")

		fileServer := http.FileServer(http.Dir("cmd/demo/resources/static"))
		router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	}
}

func newApp() *http.Server {
	addr := "0.0.0.0:8886"
	if found := os.Getenv("PORT"); found != "" {
		addr = fmt.Sprintf("0.0.0.0:%v", found)
	}
	log.Printf("Found server address %v", addr)

	opaUrl := "http://0.0.0.0:8887/v1/data/authz/allow"
	if found := os.Getenv("OPA_SERVER_URL"); found != "" {
		opaUrl = found
	}
	log.Printf("Found open policy agenet server address %v", opaUrl)

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../cmd/demo/resources")

	return App(&http.Client{}, opaUrl, addr, resourcesDirectory)
}

func main() {
	web_support.Start(newApp())
}
