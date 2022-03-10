package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/compressionsupport"
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

// todo - ignoring errors in the demo app for the moment

func (a *BasicApp) download(writer http.ResponseWriter, _ *http.Request) {
	_, file, _, _ := runtime.Caller(0)
	tar, _ := compressionsupport.TarFromPath(filepath.Join(file, "../resources/bundles"))
	_ = compressionsupport.Gzip(writer, tar)
}

func (a *BasicApp) upload(writer http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(32 << 20)
	bundleFile, _, _ := r.FormFile("bundle")
	gzip, _ := compressionsupport.UnGzip(bundleFile)
	_, file, _, _ := runtime.Caller(0)
	_ = compressionsupport.UnTarToPath(bytes.NewReader(gzip), filepath.Join(file, "../resources/bundles"))
	writer.WriteHeader(http.StatusCreated)
}

func (a *BasicApp) loadHandlers() func(router *mux.Router) {
	return func(router *mux.Router) {
		router.HandleFunc("/bundles/bundle.tar.gz", a.download).Methods("GET")
		router.HandleFunc("/bundles", a.upload).Methods("POST")
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
