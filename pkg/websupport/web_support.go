package websupport

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
)

type Model struct {
	Map map[string]interface{}
}

var resourcesDirectory string

type Options struct {
	ResourceDirectory string
}

func ModelAndView(w http.ResponseWriter, view string, data Model) error {
	views := []string{
		filepath.Join(resourcesDirectory, fmt.Sprintf("./templates/%v.gohtml", view)),
		filepath.Join(resourcesDirectory, "./templates/template.gohtml"),
	}

	base := filepath.Base(views[0]) // to match template names in ParseFiles
	return template.Must(template.New(base).Funcs(template.FuncMap{
		"capitalize": func(s string) string {
			return strings.Title(s)
		},
		"contains": func(s string, t string) bool {
			contains := strings.Contains(s, t)
			return contains
		},
		"startsWith": func(s string, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
	}).ParseFiles(views...)).Execute(w, data)
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
