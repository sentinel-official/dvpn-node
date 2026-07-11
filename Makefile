.DEFAULT_GOAL := help

GIT_COMMIT := $(shell git log -1 --format='%H' 2>/dev/null || echo "unknown")
GIT_TAG    := $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//')

comma      := ,
whitespace := $(empty) $(empty)

build_tags := netgo
ld_flags   := -s -w \
	-X github.com/sentinel-official/sentinel-go-sdk/v2/version.Commit=$(GIT_COMMIT) \
	-X github.com/sentinel-official/sentinel-go-sdk/v2/version.Tag=$(GIT_TAG)

ifeq ($(STATIC),true)
	build_tags += muslc
	ld_flags += -linkmode=external -extldflags '-Wl,-z,muldefs -static'
endif

BUILD_TAGS := $(subst $(whitespace),$(comma),$(build_tags))
LD_FLAGS   := $(ld_flags)

build_flags = -ldflags="$(LD_FLAGS)" -mod=readonly -tags="$(BUILD_TAGS)" -trimpath

GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
  GOBIN := $(shell go env GOPATH)/bin
endif

IMAGE ?= sentinel-dvpnx:latest

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the binary (./bin/sentinel-dvpnx)
	go build $(build_flags) -o ./bin/sentinel-dvpnx main.go

.PHONY: install
install: ## Install the binary into $GOBIN
	go build $(build_flags) -o "$(GOBIN)/sentinel-dvpnx" main.go

.PHONY: clean
clean: ## Remove build artifacts
	$(RM) -r ./bin ./vendor

.PHONY: test
test: ## Run tests
	go test -cover -mod=readonly -v ./...

.PHONY: benchmark
benchmark: ## Run benchmarks
	go test -bench -mod=readonly -v ./...

.PHONY: go-lint
go-lint: ## Run golangci-lint with auto-fix
	golangci-lint run --fix

.PHONY: build-image
build-image: ## Build Docker image
	docker build --compress --file Dockerfile --force-rm --tag $(IMAGE) .
