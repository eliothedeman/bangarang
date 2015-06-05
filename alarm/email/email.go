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
	Auth	*smtp.Auth
}

func NewEmail() alarm.Alarm {
	e := &Email{
		conf: &EmailConfig{},
		Auth: &Auth{},
	}
	return e
}

func (e *Email) Send(i *event.Incident) error {
	text := i.FormatDescription()
	err := smtp.SendMail(e.conf.Host+":"+strconv.Itoa(e.conf.Port), e.Auth,
		e.conf.Sender, e.conf.Recipient, []byte(text))
	return err
}

func (e *Email) ConfigStruct() interface{} {
	return e.conf
}

func (e *Email) Init(conf interface{}) error {
//	e.Auth = smtp.PlainAuth("", a.User, a.Password, a.Host)
	return nil
}
type EmailConfig struct {
	Sender		string			`json:"source_email"`
	Recipient	[]string		`json:"dest_email"`
	Host			string 			`json:"host"`
	User			string 			`json:"user"`
	Password	string 			`json:"password"`
	Port			int					`json:"port"`
}
