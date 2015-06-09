package main

import (
	"flag"

	"github.com/Sirupsen/logrus"
	_ "github.com/eliothedeman/bangarang/alarm/console"
	_ "github.com/eliothedeman/bangarang/alarm/email"
	_ "github.com/eliothedeman/bangarang/alarm/pd"
	"github.com/eliothedeman/bangarang/api"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
	_ "github.com/eliothedeman/bangarang/provider/http"
	_ "github.com/eliothedeman/bangarang/provider/tcp"
)

var (
	confFile = flag.String("conf", "/etc/bangarang/conf.json", "path main config file")
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
	tf := &logrus.TextFormatter{}
	tf.FullTimestamp = true
	logrus.SetFormatter(tf)
}

func main() {
	flag.Parse()
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
	apiServer := api.NewServer(ac.ApiPort, p, ac.Auths, ac.Hash)
	apiServer.Serve()

	<-make(chan struct{})
}
