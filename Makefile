BIN      := slimmer
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: build test lint clean

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BIN) ./cmd/slimmer

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ coverage.out

.DEFAULT_GOAL := build
