package console

import (
	"log"

	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
)

func init() {
	alarm.LoadFactory("console", NewConsole)
}

type Console struct {
}

func (c *Console) Send(e *event.Event) error {
	log.Println("%+v", *e)
	return nil
}

func (c *Console) ConfigStruct() interface{} {
	return &struct{}{}
}

func (c *Console) Init(i interface{}) error {
	return nil
}

func NewConsole() alarm.Alarm {
	return &Console{}
}
