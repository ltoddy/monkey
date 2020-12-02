.PHONY: build clean fmt install

build:
	go build -i ./...

clean:
	go clean -cache

fmt:
	gofmt -w .
	goreturns -w .

install:
	go install ./...
