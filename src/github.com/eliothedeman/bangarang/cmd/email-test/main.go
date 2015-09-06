package main

import (
	"github.com/eliothedeman/bangarang/alarm/email"
	"github.com/eliothedeman/bangarang/event"
)

func main() {
	e := email.NewEmail()
	c := e.ConfigStruct().(*email.EmailConfig)
	c.Host = "smtp.gmail.com"
	c.Port = 587
	c.Recipients = []string{"eliot.d.hedeman@gmail.com"}
	c.Sender = "bangarang"
	c.User = "bangarangtest@gmail.com"
	c.Password = "Mpc1wpOn8sUg"

	e.Init(c)
	i := &event.Incident{}
	e.Send(i)
}
