package console

import (
	"github.com/Sirupsen/logrus"

	"github.com/eliothedeman/bangarang/escalation"
	"github.com/eliothedeman/bangarang/event"
)

func init() {
	escalation.LoadFactory("console", NewConsole)
}

type Console struct {
}

func (c *Console) Send(i *event.Incident) error {

	switch i.Status {
	case event.OK:
		logrus.Info(i.FormatDescription())
	case event.WARNING:
		logrus.Warn(i.FormatDescription())
	case event.CRITICAL:
		logrus.Error(i.FormatDescription())
	}

	return nil
}

func (c *Console) ConfigStruct() interface{} {
	return &struct{}{}
}

func (c *Console) Init(i interface{}) error {
	logrus.Info("Initializing console logger.")
	return nil
}

func NewConsole() escalation.Escalation {
	return &Console{}
}
