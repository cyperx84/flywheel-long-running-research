.PHONY: build install test clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
PREFIX ?= /usr/local

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o flywheel ./cmd/flywheel/

install: build
	install -m755 flywheel $(PREFIX)/bin/flywheel

test:
	go test ./... -v

clean:
	rm -f flywheel
