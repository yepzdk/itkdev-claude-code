BINARY_NAME ?= icc
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/itk-dev/itkdev-claude-code/internal/config.version=$(VERSION)"

.PHONY: build release viewer test lint clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/icc

release:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/icc
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/icc
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/icc
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/icc

viewer:
	cd web && npm run build
	cp -r web/dist/* assets/viewer/
	cp -r assets/viewer/* internal/assets/viewer/

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

all: viewer build
