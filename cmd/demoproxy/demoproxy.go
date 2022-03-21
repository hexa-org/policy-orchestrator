package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type ProxySupport struct {
	remote  *url.URL
	reverse *httputil.ReverseProxy
	worker  workflowsupport.Worker
}

func NewProxySupport(remote *url.URL, worker workflowsupport.Worker) *ProxySupport {
	log.Println("Enabling proxy middleware.")
	return &ProxySupport{remote, httputil.NewSingleHostReverseProxy(remote), worker}
}

func (p *ProxySupport) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path := r.URL.Path; strings.Index(path, "/health") != 0 && strings.Index(path, "/_proxy") != 0 {
			log.Println("Proxying request.")
			r.URL.Host = p.remote.Host
			r.URL.Scheme = p.remote.Scheme
			r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			r.Host = p.remote.Host
			go func() {
				err := p.worker.Run([]IdentityInfo{{"someWork"}})
				if err != nil {
					log.Println("Trouble, unable to run async work.")
				}
			}()
			p.reverse.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type IdentityInfo struct {
	info string
}

type IdentityWorker struct {
}

func (i IdentityWorker) Run(work interface{}) error {
	for _, info := range work.([]IdentityInfo) {
		log.Printf("Attempting demo app proxy async work. %s\n", info.info)
	}
	return nil
}

func App(remoteUrl string, addr string, resourcesDirectory string, worker workflowsupport.Worker) *http.Server {
	foundUrl, err := url.Parse(remoteUrl)
	if err != nil {
		panic(err)
	}
	basic := NewBasicApp(foundUrl)
	server := websupport.Create(addr, basic.loadHandlers(), websupport.Options{ResourceDirectory: resourcesDirectory})
	router := server.Handler.(*mux.Router)
	proxySupport := NewProxySupport(foundUrl, worker)
	router.Use(proxySupport.Middleware)
	return server
}

type BasicApp struct {
	remote *url.URL
}

func NewBasicApp(remoteUrl *url.URL) BasicApp {
	return BasicApp{remote: remoteUrl}
}

func (a *BasicApp) dashboard(writer http.ResponseWriter, _ *http.Request) {
	_ = websupport.ModelAndView(writer, "dashboard", websupport.Model{Map: map[string]interface{}{}})
}

func (a *BasicApp) loadHandlers() func(router *mux.Router) {
	return func(router *mux.Router) {
		router.HandleFunc("/_proxy", a.dashboard).Methods("GET")

		fileServer := http.FileServer(http.Dir("cmd/demoproxy/resources/static"))
		router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	}
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Printf("Found server address %v", addr)

	remoteUrl := "http://localhost:8886"
	if found := os.Getenv("REMOTER_URL"); found != "" {
		remoteUrl = found
	}
	log.Printf("Found remoteUrl server address %v", remoteUrl)

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../resources")
	listener, _ := net.Listen("tcp", addr)
	return App(remoteUrl, listener.Addr().String(), resourcesDirectory, IdentityWorker{}), listener
}

func main() {
	websupport.Start(newApp("0.0.0.0:8890"))
}
