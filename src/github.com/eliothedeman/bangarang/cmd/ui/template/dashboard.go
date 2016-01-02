package template

import (
	html "html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
)

func init() {
	registerTemplate(&Dashboard{})
}

type Dashboard struct {
	t *html.Template
}

func (d *Dashboard) Path() string {
	return "dashboard"
}

func (d *Dashboard) Compile(src string) error {
	t := newTemplate(d.Path())
	var err error
	t, err = t.Parse(src)
	d.t = t
	return err
}

func (d *Dashboard) Compiled() bool {
	return d.t != nil
}

func (d *Dashboard) Execute(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	err := d.t.Execute(w, data)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
