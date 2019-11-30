.PHONY: build
build: format
	go build .

.PHONY: format
format:
	gofmt -w .
