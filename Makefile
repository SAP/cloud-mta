all:format test
.PHONY: format test tools

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

