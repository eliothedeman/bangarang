dir = $(shell pwd)
install:
	- go get github.com/constabulary/gb/...
	- go get -u github.com/jteeuwen/go-bindata/...
	- export PATH=$PATH:$HOME/gopath/bin
env:
	- export GOPATH=$(dir):$(dir)/vendor

test:
	- export GOPATH=$(dir):$(dir)/vendor; go test -cover -p=1 github.com/eliothedeman/bangarang/...

race:
	- export GOPATH=$(dir):$(dir)/vendor; go test -race -p=1 github.com/eliothedeman/bangarang/...

ui: env
	- cd src/github.com/eliothedeman/bangarang/cmd/ui && go-bindata -dev ./...
	- gb build github.com/eliothedeman/bangarang/cmd/ui
	- cp bin/ui src/github.com/eliothedeman/bangarang/cmd/ui/ui


testing:
	- cd src/github.com/eliothedeman/bangarang/cmd/ui && go-bindata -dev ./...
	- gb build
	- cp bin/ui src/github.com/eliothedeman/bangarang/cmd/ui/ui

generate:
	- gb generate
	- cd src/github.com/eliothedeman/bangarang/cmd/ui && go-bindata ./...

build: generate
	- gb build

deb: build 
	mkdir -p opt/bangarang
	mkdir -p etc/bangarang
	cp bin/bangarang opt/bangarang/bangarang
	cp bin/ui opt/bangarang/ui

	fpm -s dir -t deb --name bangarang -v $(shell bin/bangarang -version) etc opt
	
	rm -r opt etc
