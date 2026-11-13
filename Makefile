# roost — common tasks. `make ci` is the gate: it must pass before every commit.
BIN      := bin/roost
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS  := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)

.PHONY: build test lint fmt tidy shellcheck ci snapshot clean

build: ## Build bin/roost for this machine
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/roost

test: ## Unit tests with the race detector
	go test -race -shuffle=on ./...

lint: ## golangci-lint (includes gofmt/goimports checks)
	golangci-lint run ./...

fmt: ## Format code
	golangci-lint fmt ./...

tidy: ## Verify go.mod/go.sum are tidy
	go mod tidy && git diff --exit-code go.mod go.sum

shellcheck: ## Lint the installer
	shellcheck install.sh

ci: tidy lint test shellcheck build ## Everything CI runs

snapshot: ## Local release build of all platforms (needs goreleaser)
	goreleaser release --snapshot --clean

clean:
	rm -rf bin dist
