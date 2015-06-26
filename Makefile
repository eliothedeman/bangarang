install:
	- go get github.com/constabulary/gb/...
	- export PATH=$PATH:$GOPATH/bin

test:
	- gb test 

generate:
	- gb generate

build: generate
	- gb build github.com/eliothedeman/bangarang/cmd/bangarang

deb: build 
	mkdir -p opt/bangarang
	mkdir -p etc/bangarang
	cp bin/bangarang opt/bangarang/bangarang

	fpm -s dir -t deb --name bangarang -v $(shell bin/bangarang -version) etc opt
	
	rm -r opt etc
