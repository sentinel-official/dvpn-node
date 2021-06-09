PACKAGES := $(shell go list ./...)
VERSION := $(shell git rev-parse --short HEAD)
COMMIT := $(shell git log -1 --format='%H')

BUILD_TAGS := $(strip netgo)
LD_FLAGS := -s -w \
	-X github.com/cosmos/cosmos-sdk/version.Name=sentinel \
	-X github.com/cosmos/cosmos-sdk/version.AppName=sentinel-dvpn-node \
	-X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} \
	-X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT} \
	-X github.com/cosmos/cosmos-sdk/version.BuildTags=${BUILD_TAGS}

.PHONY: all
all: test benchmark install

.PHONY: install
install:
	go build -mod=readonly -tags="${BUILD_TAGS}" -ldflags="${LD_FLAGS}" -o ${GOPATH}/bin/sentinel-dvpn-node main.go

.PHONY: test
test:
	@go test -mod=readonly -cover ${PACKAGES}

.PHONY: benchmark
benchmark:
	@go test -mod=readonly -bench=. ${PACKAGES}
