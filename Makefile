all:format test
.PHONY: format test

format :
	go fmt ./...

lint:
	@echo "Start project linting"
	gometalinter --config=gometalinter.json ./...
	@echo "Done"

# execute general tests
test:
	 go test -v -race ./...
# check code coverage
cover:
	go test -v -coverprofile cover.out ./...
	go tool cover -html=cover.out -o cover.html
	open cover.html

