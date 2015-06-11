# bangarang [![Build Status](https://travis-ci.org/eliothedeman/bangarang.svg?branch=master)](https://travis-ci.org/eliothedeman/bangarang)

[![Join the chat at https://gitter.im/eliothedeman/bangarang](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/eliothedeman/bangarang?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
A stupid simple stream processor for monitoring applications. 

## Install
```bash
go get github.com/eliothedeman/bangarang/cmd/...
```

Or if you are on a linux system and have fpm installed.

```bash
go get github.com/eliothedeman/bangarang/cmd/...
cd $GOPATH/github.com/eliothedeman/bangarang/cmd/bangarang
./mkdeb  
```

Which will make a debian package which you can install via dpkg or your favorite package manager.

## Run
```bash
cd $GOPATH/github.com/eliothedeman/bangarang/cmd/bangarang
go build -o bangarang
./bangarang -conf="/path/to/conf.json"
```

## Configuration
bangarang uses two configurations. One main config, and a series of files that define conditions to alert on.

### Main Config
```javascript
{
	"escalations": {			// <- a list of policies to be used by the escalations
		"demo": [
			{
				"type": "console" // <- will log every event that demo is called on
			},
			{
				"type": "pager_duty", // <- creates a pagerduty event
				"key": "mytestkeyXYZ"
			},
			{
				"type": "email", // <- send an email directly via smtp
				"source_email": "ops@bigshotmonitoring.io", 
				"dest_emails": [
					"devnull@noproblemshere.com"
				],
				"host": "smtp.foo.com",
				"port": 587,
				"user": "bar",
				"password": "baz",
			}
		]
	},
	"event_providers": [
		{
			"type":"tcp", // will listen for incomming tcp connection
			"listen":"0.0.0.0:8080" // on all interfaces with port 8080
			"encoding": "msgp" // expecting msgpack encoding
		},
		{
			"type": "http", // will serve http 
			"listen": "0.0.0.0:8081", // on all interfaces with port 8081
			"encoding": "json" // expecting json encoding

		}
	],
	"escalations_dir": "/etc/bangarang/alerts/" // <- dir that holds individual alert configs
}
```

### Alert Conditions
The "escalations_dir" spesified above will be filled with seperate
.json files of alert conditions like the one below. Right now each file can only contain one alert condition. The naming of these files doens't matter, as long as it has a ".json" extension.
```javascript
{
		"match": {		// <- will pass the event on if any of the match cases are satisifed
			"service": "my.service"
		},
		"not_match": { 	// <- will pass the event on only if it doesn't match these values
			"host": "node.*host.com" 
		},
		"crit": { 		// <- the event will only be passed if the metric
						// 	   meets all of the following conditions
			"greater": 200.0,
			"less": 12.0,
			"exactly": 25.0,
			"occurences": 3 // <- will only go critical if this happens 3 times
			"escalation": "demo" // <- will be passed on to this escalation policy
		},
		"warn": {
			"greater": 200.0,
			"less": 12.0,
			"exactly": 25.0,
			"occurences": 2 // <- will go warning if it happens twice
			"escalation": "demo" // <- will be passed on to this escalation policy
		}
	}
}
```

Alerts can also be based on aggregations over a "group-by" regex pattern
```javascript
{
		"match": {		// <- will pass the event on if any of the match cases are satisifed
			"service": "my.service"
		},
		"group_by": {
			"host": "\\w+\\.(?P<deployment>\\w+)\\.\\w+" // <- will aggregate by the second element of the host name
		},
		"crit": { 		// <- the event will only be passed if the metric
						// 	   meets all of the following conditions
			"aggregation": {
				"window_length": 60 // <- number of seconds the aggregation window is open
			},
			"greater": 200.0,
			"less": 12.0,
			"exactly": 25.0,
			"occurences": 3 // <- will only go critical if this happens 3 times
			"escalation": "demo" // <- will be passed on to this escalation policy
		}
	}
}
```


## Goals
A simple stream processor for matching incoming metrics, to predefined alert conditions.

## Antigoals
Anything else.
