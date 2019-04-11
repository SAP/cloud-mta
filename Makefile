all:format test clean dir gen build-linux build-darwin build-windows copy test
.PHONY: format test tools


GOCMD=go
GOBUILD=$(GOCMD) build

# Binary names
BINARY_NAME=mta
BUILD  = $(CURDIR)/release

format :
	go fmt ./...

tools:
	@echo "Start project linting"
	curl -L https://git.io/vp6lP | bash -s -- -b $(GOPATH)/bin/ v2.0.11
	gometalinter --version
	@echo "Done"

lint:
	@echo "Start project linting"
	gometalinter --config=gometalinter.json ./...
	@echo "Done linting"

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
