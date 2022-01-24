package websupport

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

type Model struct {
	Map map[string]interface{}
}

var resourcesDirectory string

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
