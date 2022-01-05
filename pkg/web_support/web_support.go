package web_support

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
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
		filepath.Join(resourcesDirectory, fmt.Sprintf("./templates/template.gohtml")),
	}
	err := template.Must(template.ParseFiles(views...)).Execute(w, data)
	if err != nil {
		log.Printf("Unable to execute golang html templates.")
		return err
	}
	return nil
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

func Start(server *http.Server) {
	log.Printf("Starting the server.")
	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
		return
	}
}

func WaitForHealthy(server *http.Server) {
	var isLive bool
	for isLive == false {
		log.Println(fmt.Sprintf("Checking server health. %v", server.Addr))
		resp, err := http.Get(fmt.Sprintf("http://%v/health", server.Addr))
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Println("Server is healthy.")
			isLive = true
		}
	}
}

func Stop(server *http.Server) {
	log.Printf("Stopping the server.")
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Fatal(err)
		return
	}
}
