.PHONY: build clean test cover deps lint prod

build:
	go build -o bin/moria ./cmd/moria

# Compile with maximum performance optimization and smallest binary size
prod:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/moria ./cmd/moria

run: build
	./bin/moria

install:
	go install ./cmd/moria

test:
	go test ./...

cover:
	go tool cover -func=coverage.out
	go test ./... -coverprofile=coverage.out

clean:
	rm -rf bin/

GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || echo "$(HOME)/go/bin/golangci-lint")

deps:
	@if [ ! -x "$(GOLANGCI_LINT)" ]; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "golangci-lint is installed: $$($(GOLANGCI_LINT) version)"

lint: deps
	$(GOLANGCI_LINT) run ./...
