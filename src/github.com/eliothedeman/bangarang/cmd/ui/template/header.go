package template

import (
	html "html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
)

func init() {
	registerTemplate(&Header{})
}

type Header struct {
	t *html.Template
}

func (h *Header) Path() string {
	return "header"
}

func (h *Header) Compile(src string) error {
	t := newTemplate(h.Path())
	var err error
	t, err = t.Parse(src)
	h.t = t
	return err
}

func (h *Header) Compiled() bool {
	return h.t != nil
}

func (h *Header) Execute(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"logged_in": true,
	}
	err := h.t.Execute(w, data)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
