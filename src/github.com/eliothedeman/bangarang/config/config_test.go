package config

import (
	"encoding/json"
	"testing"

	_ "github.com/eliothedeman/bangarang/alarm/console"
	_ "github.com/eliothedeman/bangarang/provider/tcp"
)

func TestParseUnParse(t *testing.T) {
	v := `{
    "escalations_dir": "",
    "keep_alive_age": "25m",
    "db_path": "event.db",
    "escalations": {
    	"testing": [
    		{
    			"type": "console"
    		}
    	]
    },
    "global_policy": {
      "match": {
        "host": "\\."
      },
      "not_match": {
        "host": "unknown|shadow|telarg|sprint-17-qa|edgecast|pccw"
      },
      "group_by": null,
      "crit": null,
      "warn": null,
      "name": ""
    },
    "encoding": "json",
    "policies": {
      "settasdf": {
        "match": {
          "service": "df.free"
        },
        "not_match": {
          "host": "comcast"
        },
        "group_by": {
          "host": "^(.*)$",
          "service": "^(.*)$",
          "sub_service": "^(.*)$"
        },
        "crit": {
          "greater": null,
          "less": 5,
          "exactly": null,
          "std_dev": null,
          "escalation": "production",
          "occurences": 1,
          "window_size": 100,
          "agregation": null
        },
        "warn": null,
        "name": "settasdf"
      }
    },
    "event_providers": {
      "tcp": {
        "listen": "0.0.0.0:5555",
        "type": "tcp"
      }
    },
    "log_level": "info",
    "API_port": 8081
}`

	ac := &AppConfig{}
	err := json.Unmarshal([]byte(v), ac)
	if err != nil {
		t.Fatal(err)
	}

	if len(ac.EventProviders.Collection) != 1 {
		t.Fatal()
	}

}
