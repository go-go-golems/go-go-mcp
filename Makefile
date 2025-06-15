.PHONY: gifs

all: gifs

VERSION ?= $(shell svu)
COMMIT ?= $(shell git rev-parse --short HEAD)
DIRTY ?= $(shell git diff --quiet || echo "dirty")
LDFLAGS=-ldflags "-X main.version=$(VERSION)-$(COMMIT)-$(DIRTY)"

TAPES=$(shell ls doc/vhs/*tape)
gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

ghcr-login:
	op read "$(CR_PAT)" | docker login ghcr.io -u wesen --password-stdin


docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v2.1.0 golangci-lint run -v

lint:
	golangci-lint run -v

lintmax:
	golangci-lint run -v --max-same-issues=100

gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=.history ./...

govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

test:
	go test ./...

build:
	go generate ./...
	go build ./...

goreleaser:
	goreleaser release --skip=sign --snapshot --clean

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push --tags
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/go-go-mcp@$(shell svu current)

bump-glazed:
	go get github.com/go-go-golems/glazed@latest
	go get github.com/go-go-golems/clay@latest
	go get github.com/go-go-golems/geppetto@latest
	go get github.com/go-go-golems/parka@latest
	go mod tidy

mcp_BINARY=$(shell which mcp)
install:
	go build -o ./dist/mcp ./cmd/go-go-mcp && \
		cp ./dist/mcp $(mcp_BINARY)
