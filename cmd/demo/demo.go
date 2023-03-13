package main

import (
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hexa-org/policy-orchestrator/internal/decisionsupport"
	"github.com/hexa-org/policy-orchestrator/internal/decisionsupportproviders"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/amazonsupport"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/azuresupport"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/googlesupport"
	"github.com/hexa-org/policy-orchestrator/pkg/websupport"
)

//go:embed resources/static
var staticResources embed.FS

//go:embed resources
var resources embed.FS

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func App(session *sessions.CookieStore, amazonConfig amazonsupport.AmazonCognitoConfiguration, client HTTPClient, opaUrl string, addr string, resourcesDirectory string) *http.Server {
	basic := NewBasicApp(session, amazonConfig)
	googleSupport := googlesupport.NewGoogleSupport(session)
	amazonSupport := amazonsupport.NewAmazonSupport(client, amazonConfig, amazonsupport.AmazonCognitoClaimsParser{}, session)
	azureSupport := azuresupport.NewAzureSupport(session)
	provider := decisionsupportproviders.OpaDecisionProvider{Client: client, Url: opaUrl, Principal: "sales@hexaindustries.io"}
	opaSupport := decisionsupport.DecisionSupport{Provider: provider, Unauthorized: basic.unauthorized, Skip: []string{"/health", "/metrics", "/styles", "/images", "/bundle", "/favicon.ico"}}
	server := websupport.Create(addr, basic.loadHandlers(), websupport.Options{})
	router := server.Handler.(*mux.Router)
	router.Use(googleSupport.Middleware, amazonSupport.Middleware, azureSupport.Middleware, opaSupport.Middleware)
	return server
}

type BasicApp struct {
	session      *sessions.CookieStore
	amazonConfig amazonsupport.AmazonCognitoConfiguration
}

func NewBasicApp(session *sessions.CookieStore, amazonConfig amazonsupport.AmazonCognitoConfiguration) BasicApp {
	return BasicApp{session, amazonConfig}
}

func (a *BasicApp) dashboard(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, &resources, "dashboard", a.principalAndLogout(req))
}

func (a *BasicApp) accounting(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, &resources, "accounting", a.principalAndLogout(req))
}

func (a *BasicApp) sales(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, &resources, "sales", a.principalAndLogout(req))
}

func (a *BasicApp) humanresources(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, &resources, "humanresources", a.principalAndLogout(req))
}

func (a *BasicApp) unauthorized(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, &resources, "unauthorized", a.principalAndLogout(req))
}

func (a *BasicApp) loadHandlers() func(router *mux.Router) {
	return func(router *mux.Router) {
		router.HandleFunc("/", a.dashboard).Methods("GET")
		router.HandleFunc("/sales", a.sales).Methods("GET")
		router.HandleFunc("/accounting", a.accounting).Methods("GET")
		router.HandleFunc("/humanresources", a.humanresources).Methods("GET")

		fileServer := http.FileServer(http.FS(staticResources))
		router.PathPrefix("/").Handler(addPrefix("resources/static", fileServer))
	}
}

func addPrefix(prefix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.URL.Path = prefix + req.URL.Path
		h.ServeHTTP(rw, req)
	})
}

func (a *BasicApp) principalAndLogout(req *http.Request) websupport.Model {
	session, err := a.session.Get(req, "session")
	if err != nil {
		return websupport.Model{Map: map[string]interface{}{}}
	}
	principal := session.Values["principal"]
	if principal == nil || len(principal.([]string)) == 0 {
		return websupport.Model{Map: map[string]interface{}{}}
	}
	return websupport.Model{Map: map[string]interface{}{
		"provider_email": principal.([]string),
		"logout":         session.Values["logout"].(string),
	}}
}

func newApp(addr string) (*http.Server, net.Listener) {
	if found := os.Getenv("PORT"); found != "" {
		host, _, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", host, found)
	}
	log.Printf("Found server port %v", addr)

	if found := os.Getenv("HOST"); found != "" {
		_, port, _ := net.SplitHostPort(addr)
		addr = fmt.Sprintf("%v:%v", found, port)
	}
	log.Printf("Found server host %v", addr)

	opaUrl := "http://0.0.0.0:8887/v1/data/authz/allow"
	if found := os.Getenv("OPA_SERVER_URL"); found != "" {
		opaUrl = found
	}
	log.Printf("Found open policy agent server address %v", opaUrl)

	key := "super_private"
	if found := os.Getenv("SESSION_KEY"); found != "" {
		key = found
	}
	log.Println("Found sessions key.")

	_, file, _, _ := runtime.Caller(0)
	resourcesDirectory := filepath.Join(file, "../../../cmd/demo/resources")
	listener, _ := net.Listen("tcp", addr)
	var session = sessions.NewCookieStore([]byte(os.Getenv(key)))
	amazon := amazonsupport.AmazonCognitoConfiguration{
		Region:               os.Getenv("AWS_REGION"),
		Domain:               os.Getenv("AWS_COGNITO_USER_POOL_DOMAIN"),
		RedirectUrl:          os.Getenv("AWS_COGNITO_DOMAIN_REDIRECT_URL"),
		UserPoolId:           os.Getenv("AWS_COGNITO_USER_POOL_ID"),
		UserPoolClientId:     os.Getenv("AWS_COGNITO_USER_POOL_CLIENT_ID"),
		UserPoolClientSecret: os.Getenv("AWS_COGNITO_USER_POOL_CLIENT_SECRET"),
	}
	return App(session, amazon, &http.Client{}, opaUrl, listener.Addr().String(), resourcesDirectory), listener
}

func main() {
	websupport.Start(newApp("0.0.0.0:8886"))
}
