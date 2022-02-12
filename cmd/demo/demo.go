package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/amazonsupport"
	"github.com/hexa-org/policy-orchestrator/cmd/demo/googlesupport"
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

func App(session *sessions.CookieStore, amazonConfig amazonsupport.AmazonCognitoConfiguration, client HTTPClient, opaUrl string, addr string, resourcesDirectory string) *http.Server {
	basic := NewBasicApp(session, amazonConfig)
	googleSupport := googlesupport.NewGoogleSupport(session)
	amazonSupport := amazonsupport.NewAmazonSupport(client, amazonConfig, amazonsupport.AmazonCognitoClaimsParser{}, session)
	opaSupport := opasupport.NewOpaSupport(client, opaUrl, basic.unauthorized)
	server := websupport.Create(addr, basic.loadHandlers(), websupport.Options{ResourceDirectory: resourcesDirectory})
	router := server.Handler.(*mux.Router)
	router.Use(googleSupport.Middleware, amazonSupport.Middleware, opaSupport.Middleware)
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
	_ = websupport.ModelAndView(writer, "dashboard", a.principalAndLogout(req))
}

func (a *BasicApp) accounting(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, "accounting", a.principalAndLogout(req))
}

func (a *BasicApp) sales(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, "sales", a.principalAndLogout(req))
}

func (a *BasicApp) humanresources(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, "humanresources", a.principalAndLogout(req))
}

func (a *BasicApp) unauthorized(writer http.ResponseWriter, req *http.Request) {
	_ = websupport.ModelAndView(writer, "unauthorized", a.principalAndLogout(req))
}

func (a *BasicApp) download(writer http.ResponseWriter, _ *http.Request) {
	_, file, _, _ := runtime.Caller(0)
	opasupport.Compress(writer, filepath.Join(file, "../resources/bundles/bundle"))
}

func (a *BasicApp) loadHandlers() func(router *mux.Router) {
	return func(router *mux.Router) {
		router.HandleFunc("/", a.dashboard).Methods("GET")
		router.HandleFunc("/sales", a.sales).Methods("GET")
		router.HandleFunc("/accounting", a.accounting).Methods("GET")
		router.HandleFunc("/humanresources", a.humanresources).Methods("GET")
		router.HandleFunc("/bundles/bundle.tar.gz", a.download).Methods("GET")

		fileServer := http.FileServer(http.Dir("cmd/demo/resources/static"))
		router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	}
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
		"logout": session.Values["logout"].(string),
	}}
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
