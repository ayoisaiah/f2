APP=f2

MAKEFLAGS += --silent

.PHONY: build clean lint test
test:
	go test ./...

build:
	go build -o bin/${APP} ./cmd...

lint:
	golangci-lint run

clean:
	rm -r bin
	go clean
