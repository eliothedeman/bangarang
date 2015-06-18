package pagerduty

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type pagerdutySuccessMock struct {
	request *http.Request
}

func (pd *pagerdutySuccessMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pd.request = r
	w.Write([]byte(`{"status": "success", "message": "Event processed", "incident_key": "foobar" }`))
}

type pagerdutyErrorMock struct {
	request *http.Request
}

func (pd *pagerdutyErrorMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pd.request = r
	http.Error(w, `{"status": "invalid event", "message": "Event processed", "errors": ["foobar", "barfoo"]}`, 500)
}

// All JSON come from the Official API documentation
func TestSuccessfulRoundTrip(t *testing.T) {
	handler := &pagerdutySuccessMock{}
	server := httptest.NewServer(handler)
	Endpoint = server.URL
	defer server.Close()

	event := NewTriggerEvent("my_service_key", "my_description")
	r, statusCode, err := Submit(event)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Error("StatusCode should be 200")
	}
	if r.Status != "success" || r.Message != "Event processed" || r.IncidentKey != "foobar" {
		t.Error(r)
	}
	if event.IncidentKey != "foobar" {
		t.Error(r)
	}
}

func TestSuccessfulRoundTripIncidentKeyAlreadySet(t *testing.T) {
	handler := &pagerdutySuccessMock{}
	server := httptest.NewServer(handler)
	Endpoint = server.URL
	defer server.Close()
	event := NewTriggerEvent("my_service_key", "my_description")
	event.IncidentKey = "failure132/http"
	r, _, _ := Submit(event)
	if event.IncidentKey != "failure132/http" {
		t.Error(r)
	}
}

func TestErroredRoundTrip(t *testing.T) {
	handler := &pagerdutyErrorMock{}
	server := httptest.NewServer(handler)
	Endpoint = server.URL
	defer server.Close()
	event := NewTriggerEvent("my_service_key", "my_description")

	r, statusCode, err := Submit(event)
	if err != nil {
		t.Error(err)
	}
	if statusCode != 500 {
		t.Error("StatusCode should be 500")
	}
	if r.Status != "invalid event" || r.Message != "Event processed" || r.IncidentKey != "" {
		t.Error(r)
	}
	if r.Errors[0] != "foobar" || r.Errors[1] != "barfoo" {
		t.Error(r)
	}
}
