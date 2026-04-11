SHELL := bash

BINARY_NAME=distill
VERSION?=v0.1.0-dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

.PHONY: build test test-e2e lint clean release

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/ZouZhao321/distill/cmd.version=$(VERSION)" -o $(BINARY_NAME)$(shell go env GOEXE) .

test:
	go test ./... -v -count=1 -cover

test-e2e:
	go test ./test/e2e/ -tags=e2e -v -count=1

lint:
	go vet ./...

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

release:
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/ZouZhao321/distill/cmd.version=$(VERSION)" -o distill-linux-amd64 .
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/ZouZhao321/distill/cmd.version=$(VERSION)" -o distill-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/ZouZhao321/distill/cmd.version=$(VERSION)" -o distill-darwin-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/ZouZhao321/distill/cmd.version=$(VERSION)" -o distill-windows-amd64.exe .

tag:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
