dir = $(shell pwd)
install:
	- go get -u -v github.com/jteeuwen/go-bindata/...
	- go get -v github.com/eliothedeman/bangarang/cmd/bangarang/...
	- go get -u -v github.com/eliothedeman/randutil/...
	- go get -u github.com/glycerine/goconvey/convey
	- export PATH=$PATH:$HOME/gopath/bin

generate:
	- go generate ./...
	- cd cmd/ui && go-bindata ./...

build: generate

	- go build -o bin/bangarang github.com/eliothedeman/bangarang/cmd/bangarang 
	- go build -o bin/ui github.com/eliothedeman/bangarang/cmd/ui

test:
	go test --cover ./...

deb: build 
	mkdir -p opt/bangarang
	mkdir -p etc/bangarang
	cp bin/bangarang opt/bangarang/bangarang-server
	cp bin/ui opt/bangarang/bangarang-ui

	fpm -s dir -t deb --name bangarang -v $(shell bin/bangarang -version) etc opt
	
	rm -r opt etc bin
