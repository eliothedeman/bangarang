package main

import (
	"log"

	_ "github.com/eliothedeman/bangarang/alarm/console"
	"github.com/eliothedeman/bangarang/config"
	"github.com/eliothedeman/bangarang/pipeline"
)

func main() {
	ac, err := config.LoadConfigFile("conf.json")
	if err != nil {
		log.Fatal(err)
	}
	p := pipeline.NewPipeline(ac)
	p.Start()
	<-make(chan struct{})

}
