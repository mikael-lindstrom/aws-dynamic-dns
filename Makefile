GO        ?= go
GOOS      ?= linux
GOARCH    ?= amd64
BINDIR    := $(CURDIR)/bin
IMAGE     := lindstrom/aws-dynamic-dns
GIT_TAG   := $(shell git describe --tags --abbrev=0 2>/dev/null)

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
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 $(GO) build -o "_dist/${GOOS}-${GOARCH}/aws-dynamic-dns" .

.PHONY: docker-build
docker-build: build-cross
	cp _dist/${GOOS}-${GOARCH}/aws-dynamic-dns rootfs/
	docker build --rm -t ${IMAGE}-${GOARCH} rootfs

# usage: make docker-tag GOARCH=arm
.PHONY: docker-tag
docker-tag: docker-build
	docker tag ${IMAGE}-${GOARCH} ${IMAGE}-${GOARCH}:${GIT_TAG}

.PHONY: clean
clean:
	@rm -rf $(BINDIR) ./_dist ./rootfs/aws-dynamic-dns
