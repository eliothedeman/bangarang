package main

import (
	"flag"
	"log"

	_ "github.com/eliothedeman/bangarang/alarm/console"
	_ "github.com/eliothedeman/bangarang/alarm/pd"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

var (
	confFile = flag.String("conf", "/etc/bangarang/conf.json", "path main config file")
)

func main() {
	ac, err := config.LoadConfigFile(*confFile)
	if err != nil {
		log.Fatal(err)
	}
	p := pipeline.NewPipeline(ac)
	p.Start()
	<-make(chan struct{})
}
