package config

import (
	"testing"

	_ "github.com/eliothedeman/bangarang/provider/tcp"
)

var testConfigRaw = []byte(`
{
    "alarms": {
        "test": [
            {
                "type":"console"
            }
        ]
    },
    "event_providers": [
    	{
	    	"type": "tcp",
	    	"listen": "localhost:9999",
	    	"encoding": "json"
    	}
    ],

	"keep_alive_age": "10s",
    "escalations_dir": "alerts/"
}
`)

func TestParseConfig(t *testing.T) {
	ac, err := parseConfigFile(testConfigRaw)
	if err != nil {
		t.Error(err)
	}

	if len(*ac.EventProviders) != 1 {
		t.Fail()
	}
}
