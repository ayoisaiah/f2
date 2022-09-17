APP=f2

MAKEFLAGS += --silent

.PHONY: build clean lint test pre-commit
test:
	go test ./... --json -coverprofile=coverage.out -coverpkg .

build:
	go build -o bin/${APP} ./cmd...

lint:
	golangci-lint run

pre-commit:
	pre-commit run

clean:
	rm -r bin
	go clean
