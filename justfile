APP := "f2"
toolprefix := "go tool -modfile=" + justfile_directory() + "/tools.mod"
toolsmod := "-modfile=" + justfile_directory() + "/tools.mod"

# Run all tests
test filter='.*':
	@go test ./... -coverprofile=coverage.out -coverpkg=. -json -run={{filter}} | {{toolprefix}} gotestfmt -hide 'empty-packages'

[no-cd]
test-pkg filter='.*':
  @go test ./... -json -coverprofile=coverage.out -coverpkg=. -run={{filter}} | {{toolprefix}} gotestfmt -hide 'empty-packages'

[no-cd]
update-golden filter='.*':
  @go test ./... -update -json -run={{filter}} | {{toolprefix}} gotestfmt

alias i := install

install:
  @go mod download
  @go mod download {{toolsmod}}

add pkg:
  @go get -u {{pkg}}

add-tool tool:
  @go get {{toolsmod}} -tool {{tool}}

show-updates:
  @echo "# Main dependencies"
  @go list -u -f '{{"{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}"}}' -m all
  @echo ""
  @echo "# Tool dependencies"
  @go list {{toolsmod}} -u -f '{{"{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}"}}' -m all

update-deps:
  @go get -u ./...

update-tools:
  @go get {{toolsmod}} -u ./...

update: && update-deps update-tools

tools:
  @go tool {{toolsmod}}

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
