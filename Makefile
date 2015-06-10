BUILD=go build -a
CWD=$(shell pwd)
IMAGE_NAME=bangarang

make:
	go generate ./...
	$(BUILD) -o bin/bangarang github.com/eliothedeman/bangarang/cmd/bangarang

test:
	go test -p=1 ./...

all: linux osx windows

linux:
	GOOS="linux" GOARCH="amd64" $(BUILD) -o bin/linux/bangarang github.com/eliothedeman/bangarang/cmd/bangarang

osx:
	GOOS="darwin" GOARCH="amd64" $(BUILD) -o bin/darwin/bangarang github.com/eliothedeman/bangarang/cmd/bangarang

windows:
	GOOS="windows" GOARCH="amd64" $(BUILD) -o bin/windows/bangarang github.com/eliothedeman/bangarang/cmd/bangarang


# docker stuff
build:
	docker build --no-cache -t $(IMAGE_NAME) .

start: 
	cwd=$(pwd)
	docker run -v $(CWD)/alerts:/etc/bangarang/alerts -v $(CWD)/conf.json:/etc/bangarang/conf.json -p 8081:8081 -p 5555:5555 -p 5556:5556 --name $(IMAGE_NAME) -d $(IMAGE_NAME)

start-no-d: 
	cwd=$(pwd)
	docker run -v $(CWD)/alerts:/etc/bangarang/alerts -v $(CWD)/conf.json:/etc/bangarang/conf.json -p 8081:8081 -p 5555:5555 -p 5556:5556 --name $(IMAGE_NAME) $(IMAGE_NAME)

stop:
	docker kill $(IMAGE_NAME)

clean: stop
	docker rm $(IMAGE_NAME)
