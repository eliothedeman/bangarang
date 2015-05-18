package main

import (
	_ "github.com/eliothedeman/bangarang/alarm/console"
	"github.com/eliothedeman/bangarang/pipeline"
)

func main() {
	tcp := 8080
	p := pipeline.NewPipeline(&tcp, nil)
	p.Start()
	<-make(chan struct{})

}
