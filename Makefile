GO        ?= go
BINDIR    := $(CURDIR)/bin

.PHONY: all
all: build

.PHONY: build
build:
	GOBIN=$(BINDIR) $(GO) install .

.PHONY: format
format:
	find . -type f -name "*.go" | xargs gofmt -s -w

# usage: make build-cross GOOS=linux GOARCH=arm
.PHONY: build-cross
build-cross:
	CGO_ENABLED=0 $(GO) build -o "_dist/${GOOS}-${GOARCH}/aws-dynamic-dns" .

.PHONY: clean
clean:
	@rm -rf $(BINDIR) ./_dist
