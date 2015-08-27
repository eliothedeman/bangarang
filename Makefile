dir = $(shell pwd)
install:
	- go get github.com/constabulary/gb/...
	- go get -u github.com/jteeuwen/go-bindata/...
	- go get -u github.com/pquerna/ffjson
	- export PATH=$PATH:$HOME/gopath/bin

test:
	- export GOPATH=$(dir):$(dir)/vendor; go test -p=1 github.com/eliothedeman/bangarang/...


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
