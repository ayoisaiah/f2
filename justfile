APP := "f2"

test:
	@go test ./... -json | gotestfmt -hide 'empty-packages'

test-race:
	@go test ./... -json -race | gotestfmt -hide 'empty-packages'

[no-cd]
test-pkg filter='.*':
  @go test ./... -json -coverprofile=coverage.out -coverpkg=. -run={{filter}} | gotestfmt -hide 'empty-packages'

[no-cd]
update-golden filter='.*':
  @go test ./... -update -json -race -run={{filter}} | gotestfmt

build:
	@go build -o bin/{{APP}} ./cmd...

build-win:
	@go build -o bin/{{APP}}.exe ./cmd...

lint:
	@golangci-lint run ./...

[no-cd]
lint-pkg:
	@golangci-lint run ./...

pre-commit:
	@pre-commit run

clean:
	@rm -r bin
	@go clean

sloc:
  tokei
