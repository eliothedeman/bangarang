package template

import (
	html "html/template"
	"log"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/api/client"
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
	c := client.NewClientWithAuthToken(tokenFromRequest(r))
	u, err := c.GetSelf()
	log.Println(u)

	data := map[string]interface{}{
		"logged_in": err == nil,
		"user":      u,
	}
	err = h.t.Execute(w, data)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
