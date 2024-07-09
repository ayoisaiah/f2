APP := "f2"

test:
	@go test ./... --json -coverprofile=coverage.out -coverpkg .

[no-cd]
test-pkg:
    @go test --json -coverprofile=coverage.out -coverpkg=../ | gotestfmt

[no-cd]
update-golden:
    @go test --update --json | gotestfmt

build:
	@go build -o bin/{{APP}} ./cmd...

lint:
	@golangci-lint run ./...

pre-commit:
	@pre-commit run

clean:
	@rm -r bin
	@go clean

sloc:
  tokei
