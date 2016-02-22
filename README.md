# bangarang [![Build Status](https://travis-ci.org/eliothedeman/bangarang.svg?branch=master)](https://travis-ci.org/eliothedeman/bangarang) [![Join the chat at https://gitter.im/eliothedeman/bangarang](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/eliothedeman/bangarang?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
A stupid simple stream processor for monitoring applications. 

![Imgur](http://i.imgur.com/oUQ4RDC.png)

## Install

To build and run bangarang from source, run...

### install dependancies
```bash
make install
```

and build bangarang
```bash
make build
```

## Run

### Start the server
```bash
bin/bangarang -conf="/where/you/want/the/conf.db"
```

### Start the ui server
```bash
bin/ui -api="localhost:8081" -l=":9090"
```
Then simply open a browser to localhost:9090 and begin configuration


## Client libraries
As of now, there is only a go client library, which can be found [here](https://github.com/eliothedeman/go-bangarang)

## Development
Bangarang is still under heavy development, but a 1.0 release is coming soon. The stream processor is mostly feature complete, with most of the changes currently being made to the UI and rest API.

Pull request and questions are gladly welcomed.

## Goals
A simple stream processor for matching incoming metrics, to predefined alert conditions.

## Antigoals
Anything else.
