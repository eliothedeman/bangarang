package template

import (
	"fmt"
	html "html/template"
	"net/http"
)

var templates = make(map[string]Template)

func registerTemplate(t Template) {
	templates[t.Path()] = t
}

func newTemplate(name string) *html.Template {
	t := html.New(name)
	t.Delims("[[", "]]")

	return t
}

func Get(path string) Template {
	t, ok := templates[path]
	if !ok {
	}
	return t
}

type Template interface {
	Path() string
	Compile(src string) error
	Compiled() bool
	Execute(w http.ResponseWriter, r *http.Request)
}

func templatePath(path string) string {
	return fmt.Sprintf("template/%s.gohtml")
}
