BINARY_NAME=distill
VERSION?=0.1.0-mvp

.PHONY: build test lint clean release

build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o $(BINARY_NAME)$(shell go env GOEXE) .

test:
	go test ./... -v -count=1 -cover

lint:
	go vet ./...

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

release:
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o distill-linux-amd64 .
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o distill-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -o distill-darwin-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o distill-windows-amd64.exe .

tag:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
