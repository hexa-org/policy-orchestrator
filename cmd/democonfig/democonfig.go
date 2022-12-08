package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/hexa-org/policy-orchestrator/internal/compressionsupport"
	"io/fs"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
)

func App(addr string) *http.Server {
	basic := NewBasicApp()
	return websupport.Create(addr, basic.loadHandlers(), websupport.Options{})
}

type BasicApp struct {
}

func NewBasicApp() BasicApp {
	return BasicApp{}
}

// todo - ignoring errors in the demo app for the moment

func (a *BasicApp) download(writer http.ResponseWriter, _ *http.Request) {
	_, file, _, _ := runtime.Caller(0)
	bundles := filepath.Join(file, "../resources/bundles")
	tar, _ := compressionsupport.TarFromPath(fmt.Sprintf("%s/%s", bundles, a.latest(bundles)))
	_ = compressionsupport.Gzip(writer, tar)
}

func (a *BasicApp) latest(dir string) string {
	available := a.available(dir)
	return available[0].Name()
}

func (a *BasicApp) available(dir string) []fs.FileInfo {
	available := make([]fs.FileInfo, 0)
	_ = fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
		info, _ := d.Info()
		if info.Name() == "." {
			return nil
		}
		available = append(available, info)
		return fs.SkipDir
	})
	sort.Slice(available, func(i, j int) bool {
		return available[i].ModTime().Unix() > available[j].ModTime().Unix()
	})
	return available
}

func (a *BasicApp) upload(writer http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(32 << 20)
	bundleFile, _, _ := r.FormFile("bundle")
	gzip, _ := compressionsupport.UnGzip(bundleFile)
	_, file, _, _ := runtime.Caller(0)
	rand.Seed(time.Now().UnixNano())
	path := filepath.Join(file, fmt.Sprintf("../resources/bundles/.bundle-%d", rand.Uint64()))
	_ = compressionsupport.UnTarToPath(bytes.NewReader(gzip), path)
	writer.WriteHeader(http.StatusCreated)
}

func (a *BasicApp) reset(writer http.ResponseWriter, r *http.Request) {
	_, file, _, _ := runtime.Caller(0)
	bundles := filepath.Join(file, "../resources/bundles")
	for _, available := range a.available(bundles) {
		if strings.Index(available.Name(), ".bundle") == 0 {
			path := filepath.Join(bundles, available.Name())
			err := os.RemoveAll(path)
			if err != nil {
				log.Printf("Unable to remove bundle %v", path)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func (a *BasicApp) loadHandlers() func(router *mux.Router) {
	return func(router *mux.Router) {
		router.HandleFunc("/bundles/bundle.tar.gz", a.download).Methods("GET")
		router.HandleFunc("/bundles", a.upload).Methods("POST")
		router.HandleFunc("/reset", a.reset).Methods("GET")
	}
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Printf("Found server address %v", addr)

	if found := os.Getenv("HOST"); found != "" {
		_, port, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", found, port)
	}
	log.Printf("Found server host %v", addr)

	listener, _ := net.Listen("tcp", addr)
	app := App(listener.Addr().String())

	serverCertPath := os.Getenv("SERVER_CERT")
	serverKeyPath := os.Getenv("SERVER_KEY")
	if serverKeyPath != "" && serverCertPath != "" {
		key, err := os.ReadFile(serverKeyPath)
		if err != nil {
			panic(fmt.Sprintf("invalid SERVER_KEY path: %s", err))
		}
		cert, err := os.ReadFile(serverCertPath)
		if err != nil {
			panic(fmt.Sprintf("invalid SERVER_CERT path: %s", err))
		}
		pair, err := tls.X509KeyPair(cert, key)
		if err != nil {
			panic(fmt.Sprintf("invalid cert/key pair: %s", err))
		}
		app.TLSConfig = &tls.Config{
			// todo - tls client auth? Should we require client cert verification?
			Certificates: []tls.Certificate{pair},
		}
	}

	return app, listener
}

func main() {
	app, listener := newApp("0.0.0.0:8889")

	if os.Getenv("SERVER_KEY") != "" && os.Getenv("SERVER_CERT") != "" {
		websupport.StartWithTLS(app, listener)
		return
	}

	websupport.Start(app, listener)
}
