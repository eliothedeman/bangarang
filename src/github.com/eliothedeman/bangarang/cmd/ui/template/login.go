package template

import (
	html "html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
)

func init() {
	registerTemplate(&Login{})
}

type Login struct {
	t *html.Template
}

func (l *Login) Path() string {
	return "login"
}

func (l *Login) Compile(src string) error {
	t := newTemplate(l.Path())
	var err error
	t, err = t.Parse(src)
	l.t = t
	return err
}

func (l *Login) Compiled() bool {
	return l.t != nil
}

func (l *Login) Execute(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	err := l.t.Execute(w, data)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
