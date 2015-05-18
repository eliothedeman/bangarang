package pd

import (
	"github.com/eliothedeman/bangarang/alarm"
	"github.com/eliothedeman/bangarang/event"
	"github.com/marcw/pagerduty"
)

func init() {
	alarm.LoadFactory("pager_duty", NewPagerduty)
}

type PagerDuty struct {
	conf *PagerDutyConfig
}

func NewPagerduty() alarm.Alarm {
	p := &PagerDuty{
		conf: &PagerDutyConfig{},
	}
	return p
}

func (p *PagerDuty) ConfigStruct() interface{} {
	return p.conf
}

func (p *PagerDuty) Init(conf interface{}) error {
	return nil
}

func (p *PagerDuty) Send(e *event.Event) error {
	var pdPevent *pagerduty.Event
	switch e.Status {
	case event.CRITICAL, event.WARNING:
		pdPevent = pagerduty.NewTriggerEvent(p.conf.Key, e.FormatDescription())
	case event.OK:
		pdPevent = pagerduty.NewResolveEvent(p.conf.Key, e.FormatDescription())
	}
	pdPevent.IncidentKey = e.IndexName()

	_, _, err := pagerduty.Submit(pdPevent)
	return err
}

type PagerDutyConfig struct {
	Subdomain string `json:"subdomain"`
	Key       string `json:"key"`
}
