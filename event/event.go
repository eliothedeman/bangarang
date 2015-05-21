package event

import "fmt"

const (
	OK = iota
	WARNING
	CRITICAL
)

//go:generate ffjson $GOFILE

type Event struct {
	Host       string            `json:"host"`
	Service    string            `json:"service"`
	SubService string            `json:"sub_type"`
	Metric     float64           `json:"metric"`
	Occurences int               `json:"occurences"`
	Tags       map[string]string `json:"tags"`
	Status     int               `json:"status"`
	LastEvent  *Event            `json:"last_event,omitempty"`
	indexName  string            `json:"index_name"`
}

func (e *Event) IndexName() string {
	if len(e.indexName) == 0 {
		e.indexName = fmt.Sprintf("%s:%s:%s", e.Host, e.Service, e.SubService)
	}
	return e.indexName
}

func (e *Event) StatusChanged() bool {
	if e.LastEvent == nil {
		return e.Status != OK
	}

	return !(e.LastEvent.Status == e.Status)
}

func status(code int) string {
	switch code {
	case WARNING:
		return "warning"
	case CRITICAL:
		return "critical"
	default:
		return "ok"
	}
}

func (e *Event) FormatDescription() string {
	return fmt.Sprintf("%s! %s.%s on %s is %.2f", status(e.Status), e.Service, e.SubService, e.Host, e.Metric)
}
