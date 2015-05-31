BUILD=go build -a

make:
	$(BUILD) -o bin/bangarang github.com/eliothedeman/bangarang/cmd/bangarang

test:
	go test ./...

all: linux osx windows

linux:
	GOOS="linux" GOARCH="amd64" $(BUILD) -o bin/linux/bangarang github.com/eliothedeman/bangarang/cmd/bangarang

osx:
	GOOS="darwin" GOARCH="amd64" $(BUILD) -o bin/darwin/bangarang github.com/eliothedeman/bangarang/cmd/bangarang

windows:
	GOOS="windows" GOARCH="amd64" $(BUILD) -o bin/windows/bangarang github.com/eliothedeman/bangarang/cmd/bangarang
