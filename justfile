APP := "f2"
toolprefix := "go tool -modfile=tools.mod"

# Run all tests
test:
	@go test ./... -coverprofile=coverage.out -coverpkg=. -json | {{toolprefix}} gotestfmt -hide 'empty-packages'

[no-cd]
test-pkg filter='.*':
  @go test ./... -json -coverprofile=coverage.out -coverpkg=. -run={{filter}} | {{toolprefix}} gotestfmt -hide 'empty-packages'

[no-cd]
update-golden filter='.*':
  @go test ./... -update -json -run={{filter}} | {{toolprefix}} gotestfmt

install:
  @go mod download
  @go mod download -modfile=tools.mod

add pkg:
  @go get -u {{pkg}}

add-tool tool:
  @go get -modfile=tools.mod -tool {{tool}}

show-updates:
  @go list -u -f '{{"{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}"}}' -m all
  @go list -modfile=tools.mod -u -f '{{"{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}"}}' -m all

update-deps:
  @go get -u ./...

update-tools:
  @go get -modfile=tools.mod -u ./...

update: && update-deps update-tools

tools:
  @go tool -modfile=tools.mod

build:
	@go build -o bin/{{APP}} ./cmd...

build-win:
	@go build -o bin/{{APP}}.exe ./cmd...

lint:
	@{{toolprefix}} golangci-lint run ./...

pre-commit:
	@pre-commit run

clean:
	@rm -r bin
	@go clean

scc:
  @{{toolprefix}} scc
