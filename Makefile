VERSION ?= 0.4.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X github.com/rakshithgoud453/myvault/internal/cli.Version=$(VERSION) -w -s

.PHONY: all build test clean dist

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o myvault ./cmd/myvault

install: build
	cp ./myvault ~/.local/bin/myvault
	@if [ "$$(uname)" = "Darwin" ]; then codesign -f -s - ~/.local/bin/myvault 2>/dev/null || true; fi

test:
	go test -v ./...

clean:
	rm -rf myvault dist/

dist: test
	mkdir -p dist
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/myvault-darwin-arm64 ./cmd/myvault
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/myvault-darwin-amd64 ./cmd/myvault
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/myvault-linux-amd64 ./cmd/myvault
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/myvault-linux-arm64 ./cmd/myvault
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/myvault-windows-amd64.exe ./cmd/myvault
	@echo "Build complete. Artifacts in dist/:"
	@ls -lh dist/
