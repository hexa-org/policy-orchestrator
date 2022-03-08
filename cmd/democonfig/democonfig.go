package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/cmd/democonfig/gzipsupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func App(addr string, resourcesDirectory string) *http.Server {
	basic := NewBasicApp()
	return websupport.Create(addr, basic.loadHandlers(), websupport.Options{ResourceDirectory: resourcesDirectory})
}

type BasicApp struct {
}

func NewBasicApp() BasicApp {
	return BasicApp{}
}

func (a *BasicApp) download(writer http.ResponseWriter, _ *http.Request) {
	_, file, _, _ := runtime.Caller(0)
	gzipsupport.Compress(writer, filepath.Join(file, "../resources/bundles/bundle"))
}

func (a *BasicApp) loadHandlers() func(router *mux.Router) {
	return func(router *mux.Router) {
		router.HandleFunc("/bundles/bundle.tar.gz", a.download).Methods("GET")
	}
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Printf("Found server address %v", addr)

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../cmd/demo/resources")
	listener, _ := net.Listen("tcp", addr)
	return App(listener.Addr().String(), resourcesDirectory), listener
}

func main() {
	websupport.Start(newApp("0.0.0.0:8889"))
}
