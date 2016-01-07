package template

import (
	html "html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/api/client"
	"github.com/eliothedeman/bangarang/event"
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
	funcs := html.FuncMap{
		"statusColor": func(status int) string {
			color := ""
			switch status {
			case event.OK:
				color = "green"
			case event.WARNING:
				color = "yellow"
			default:
				color = "red"
			}

			return color

		},
		"statusCode": event.Status,
	}
	t = t.Funcs(funcs)
	var err error
	t, err = t.Parse(src)
	d.t = t
	return err
}

func (d *Dashboard) Compiled() bool {
	return d.t != nil
}

func (d *Dashboard) Execute(w http.ResponseWriter, r *http.Request) {
	c := client.NewClientWithAuthToken(tokenFromRequest(r))
	i, err := c.GetIncidents(0, 0)
	if err != nil {
		logrus.Error(err)
	}

	data := map[string]interface{}{
		"incidents": i,
	}
	err = d.t.Execute(w, data)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
