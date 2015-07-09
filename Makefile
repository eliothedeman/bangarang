dir = $(shell pwd)
install:
	- go get github.com/constabulary/gb/...
	- export PATH=$PATH:$GOPATH/bin

test:
	- gb test 

testing: generate
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

	fpm -s dir -t deb --name bangarang -v $(shell bin/bangarang -version) etc opt
	
	rm -r opt etc
