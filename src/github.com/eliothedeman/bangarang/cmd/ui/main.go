package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/eliothedeman/bangarang/cmd/ui/template"
	"github.com/eliothedeman/bangarang/config"
)

type Server struct {
}

func getContentType(fileName string) string {
	n := strings.Split(fileName, ".")
	if len(n) == 0 {
		return "application/json"
	}
	switch n[len(n)-1] {
	case "ico":
		return "image/png"
	case "png":
		return "image/png"
	case "jgp":
		return "image/jpeg"
	case "svg":
		return "image/svg"
	case "css":
		return "text/css"
	case "html", "htm":
		return "text/html"
	case "js":
		return "application/javascript"
	default:
		return "application/json"
	}
	return ""
}

func templateName(name string) string {
	return fmt.Sprintf("template/%s.gohtml", name)
}

func wrapHeader(w http.ResponseWriter, r *http.Request) {
	buff, err := Asset(templateName("header"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Write(buff)
}

func wrapFooter(w http.ResponseWriter, r *http.Request) {
	buff, err := Asset(templateName("footer"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Write(buff)
}

func compileTemplate(t template.Template) error {

	if !*dev {
		// stop short
		if t.Compiled() {
			return nil
		}
	}

	buff, err := Asset(templateName(t.Path()))
	if err != nil {
		return err
	}

	return t.Compile(string(buff))
}

func (s *Server) GetSelf(r *http.Request) (*config.User, error) {
	r.Method = "GET"
	r.URL.Path = "api/user"
	r.URL.Host = *api_host
	r.URL.Scheme = "http"
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	u := &config.User{}
	err = json.Unmarshal(buff, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	path := r.URL.Path
	logrus.Infof("Serving %s", path)

	if strings.HasPrefix(path, "/api") {
		logrus.Infof("Relaying api call %s", path)
		r.RequestURI = ""
		r.URL.Scheme = "http"
		r.URL.Host = *api_host
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(resp.StatusCode)
		w.Header().Add("Content-Type", "application/json")
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	// serve the index.html if no path is given
	if path == "/" || path == "" {
		path = "dashboard"
	} else {
		path = path[1:]
	}

	// handle if this is a template
	t := template.Get(path)
	if t != nil {

		// check to see if the user is loged in
		u, err := s.GetSelf(r)

		// if not, redirect to the login page
		if err != nil && path != "/login" {
			r.URL.Path = "/login"
			r.Method = "GET"
			s.ServeHTTP(w, r)
			return
		}
		log.Println(u)

		w.Header().Add("Content-Type", "text/html")
		wrapHeader(w, r)
		err = compileTemplate(t)
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t.Execute(w, r)
		wrapFooter(w, r)
		return
	} else {
		w.Header().Add("Content-Type", getContentType(path))

		// regular assets
		buff, err := Asset(path)
		if err != nil {
			logrus.Error(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Write(buff)
	}

}

var (
	api_host = flag.String("api", "localhost:8081", "")
	listen   = flag.String("l", ":9090", "serve http on")
	root     = flag.String("r", ".", "root dir for assets")
	dev      = flag.Bool("dev", false, "run in developer mode")
	rootDir  = "."
)

func main() {
	flag.Parse()
	rootDir = *root
	s := &Server{}
	logrus.SetLevel(logrus.InfoLevel)

	http.ListenAndServe(*listen, s)
}
