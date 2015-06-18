install:
	- go get github.com/constabulary/gb/...
	- export PATH=$PATH:$GOPATH/bin

test:
	- gb test 

