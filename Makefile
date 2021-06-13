all:format test clean dir gen build-linux build-darwin build-darwin-arm build-windows copy test
.PHONY: format test tools


GOCMD=go
GOBUILD=$(GOCMD) build
GOLANGCI_VERSION = 1.21.0

# Binary names
BINARY_NAME=mta
BUILD  = $(CURDIR)/release

format :
	go fmt ./...

tools:
	@echo "download golangci-lint"
	curl -sLO https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_VERSION}/golangci-lint-${GOLANGCI_VERSION}-linux-amd64.tar.gz
	tar -xzvf golangci-lint-${GOLANGCI_VERSION}-linux-amd64.tar.gz
	cp golangci-lint-${GOLANGCI_VERSION}-linux-amd64/golangci-lint $(GOPATH)/bin
	@echo "done"

lint:
	@echo "Start project linting"
	golangci-lint run --config .golangci.yml
	@echo "done linting"

# execute general tests
test:
	 go test -v -race ./...

# check code coverage
cover:
	go test -v -coverprofile cover.out ./...
	go tool cover -html=cover.out -o cover.html
	open cover.html

clean:
	rm -rf $(BUILD)

dir:
	mkdir $(BUILD)

gen:
	go generate


# build for each platform
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o release/$(BINARY_NAME)_linux -v

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o release/$(BINARY_NAME) -v

build-darwin-arm:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -o release/$(BINARY_NAME)_darwin_arm -v

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o release/$(BINARY_NAME)_windows -v

# use for local development - > copy the new bin to go/bin path to use new compiled version
copy:
ifeq ($(OS),Windows_NT)
	cp $(CURDIR)/release/$(BINARY_NAME)_windows $(GOPATH)/bin/$(BINARY_NAME).exe
else
	cp $(CURDIR)/release/$(BINARY_NAME) $(GOPATH)/bin/
	cp $(CURDIR)/release/$(BINARY_NAME) $~/usr/local/bin/
endif
