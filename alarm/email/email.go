package email

import (
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
	"net/smtp"
)

func init() {
	alarm.LoadFactory("email", NewEmail)
}

type Email struct {
	conf *EmailConfig
}

func NewEmail() alarm.Alarm {
	e := &Email{
		conf: &EmailConfig{},
	}
	return e
}

func (e *Email) Send(i *event.Incident) error {
	text := i.FormatDescription()
	err := smtp.SendMail(e.conf.Server.Host, e.conf.Server.Auth,
		e.conf.Sender, e.conf.Recipient, []byte(text))
	return err
}

func (e *Email) ConfigStruct() interface{} {
	return e.conf
}

func (p *Email) Init(conf interface{}) error {
	return nil
}

func (a *SMTPServer) Init(conf interface{}) error {
	a.Auth = smtp.PlainAuth(a.Identity, a.User, a.Password, a.Host)
	return nil
}

type EmailConfig struct {
	Sender    string     `json:"source_email"`
	Recipient []string   `json:"dest_email"`
	Server    SMTPServer `json:"server"`
}

type SMTPServer struct {
	Host     string `json:"host"`
	Identity string `json:"identity"`
	User     string `json:"user"`
	Password string `json:"password"`
	Auth     smtp.Auth
}
