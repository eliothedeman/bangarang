package event

import "fmt"

const (
	OK = iota
	WARNING
	CRITICAL
)

//go:generate ffjson $GOFILE
//go:generate msgp $GOFILE

type Event struct {
	Host       string            `json:"host" msg:"host"`
	Service    string            `json:"service" msg:"service"`
	SubService string            `json:"sub_type" msg:"sub_service"`
	Metric     float64           `json:"metric" msg:"metric"`
	Occurences int               `json:"occurences" msg:"occurences"`
	Tags       map[string]string `json:"tags" msg:"tags"`
	Status     int               `json:"status" msg:"status"`
	IncidentId *int64            `json:"incident,omitempty" msg:"incident_id"`
	indexName  string
}

func (e *Event) IndexName() string {
	if len(e.indexName) == 0 {
		e.indexName = e.Host + e.Service + e.SubService
	}
	return e.indexName
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
