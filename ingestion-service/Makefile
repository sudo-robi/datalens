.PHONY: test test-coverage build

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

build:
	CGO_ENABLED=0 go build -o server .
