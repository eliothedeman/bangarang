package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	_ "github.com/eliothedeman/bangarang/alarm/console"
	_ "github.com/eliothedeman/bangarang/alarm/email"
	_ "github.com/eliothedeman/bangarang/alarm/influxdb"
	_ "github.com/eliothedeman/bangarang/alarm/pd"
	"github.com/eliothedeman/bangarang/api"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
	_ "github.com/eliothedeman/bangarang/provider/http"
	_ "github.com/eliothedeman/bangarang/provider/tcp"
	"github.com/eliothedeman/bangarang/ui"
)

var (
	confFile = flag.String("conf", "/etc/bangarang/conf.json", "path main config file")
	dev      = flag.Bool("dev", false, "puts bangarang in a dev testing mode")
	version  = flag.Bool("version", false, "display the version of this binary")
)

const (
	VERSION = "0.3.2"
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
	tf := &logrus.TextFormatter{}
	tf.FullTimestamp = true
	tf.ForceColors = true
	logrus.SetFormatter(tf)
}

func handleSigs() {
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Kill, os.Interrupt)

	done := <-stop
	logrus.Fatal(done.String())
}

func main() {
	flag.Parse()

	// display the current version and exit
	if *version {
		fmt.Print(VERSION)
		os.Exit(0)
	}

	logrus.Infof("Loading config file %s", *confFile)
	ac, err := config.LoadConfigFile(*confFile)
	if err != nil {
		logrus.Fatal(err)
	}

	if ac.LogLevel == "" {
		ac.LogLevel = config.DEFAULT_LOG_LEVEL
	}

	ll, err := logrus.ParseLevel(ac.LogLevel)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.SetLevel(ll)
	}

	logrus.Infof("Starting processing pipeline with %d policie(s)", len(ac.Policies))
	// create and start up a new pipeline
	p := pipeline.NewPipeline(ac)
	p.Start()

	logrus.Infof("Serving the http api on port %d", 8081)
	// create and start a new api server
	apiServer := api.NewServer(ac.ApiPort, p, ac.Auths)
	go apiServer.Serve()

	// start the ui
	go func() {
		uiServer := &ui.Server{}
		err := http.ListenAndServe(":9090", uiServer)
		if err != nil {
			log.Fatal(err)
		}
	}()
	handleSigs()
}
