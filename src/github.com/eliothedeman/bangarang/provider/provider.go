package provider

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eliothedeman/bangarang/event"
)

var (
	// holds factories for all loaded EventProviders
	EVENT_PROVIDER_FACTORIES = map[string]EventProviderFactory{}
	PROVIDER_TYPE_NOT_FOUND  = errors.New("A provider type must be spesified")
)

type INVALID_PROVIDER_TYPE string

func (i INVALID_PROVIDER_TYPE) Error() string {
	return fmt.Sprintf("Unknown provider type: %s", i)
}

type EventProviderCollection []EventProvider

func (e *EventProviderCollection) UnmarshalJSON(buff []byte) error {
	typer := struct {
		Type string `json:"type"`
	}{}

	// turn the buff into an array of buffs
	eps := make([]json.RawMessage, 0)
	err := json.Unmarshal(buff, &eps)

	for _, b := range eps {
		typer.Type = ""

		// get the type of the provider
		err = json.Unmarshal(b, &typer)
		if err != nil {
			return err
		}

		// if no type was found, error out
		if typer.Type == "" {
			return PROVIDER_TYPE_NOT_FOUND
		}

		p := GetEventProvider(typer.Type)
		if p == nil {
			return INVALID_PROVIDER_TYPE(typer.Type)
		}

		conf := p.ConfigStruct()
		err = json.Unmarshal(b, conf)
		if err != nil {
			return err
		}

		err = p.Init(conf)
		if err != nil {
			return err
		}

		*e = append(*e, p)
	}
	return nil
}

// Load a given event provider factory
func LoadEventProviderFactory(name string, f EventProviderFactory) {
	EVENT_PROVIDER_FACTORIES[name] = f
}

// Get an event provider by name
func GetEventProvider(name string) EventProvider {
	f := EVENT_PROVIDER_FACTORIES[name]
	return f()
}

// Provides an interface for injesting events from an outside service
type EventProvider interface {
	Start(dst chan *event.Event)
	ConfigStruct() interface{}
	Init(interface{}) error
}

// create and return a new EventProvider
type EventProviderFactory func() EventProvider
