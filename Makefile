SRCS := $(shell find . -type d -name vendor -prune -o -type f -name "*.go" -print)

.PHONY: build
build: format
	go build .

.PHONY: format
format: imports
	gofmt -w .

.PHONY: imports
imports:
	-@goimports -w $(SRCS)
