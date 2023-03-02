PACKAGES := $(shell go list ./...)
VERSION := $(shell git describe --tags | sed 's/^v//' | cut -d - -f 1-2)
COMMIT := $(shell git log -1 --format='%H')

BUILD_TAGS := $(strip netgo)
LD_FLAGS := -s -w \
	-X github.com/cosmos/cosmos-sdk/version.Name=sentinel \
	-X github.com/cosmos/cosmos-sdk/version.AppName=sentinelnode \
	-X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} \
	-X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT} \
	-X github.com/cosmos/cosmos-sdk/version.BuildTags=${BUILD_TAGS}

.PHONY: benchmark
benchmark:
	@go test -mod=readonly -v -bench= ${PACKAGES}

.PHONY: clean
clean:
	rm -rf ./bin ./build ./vendor

.PHONY: install
install:
	go build -mod=readonly -tags="${BUILD_TAGS}" -ldflags="${LD_FLAGS}" \
		-o ${GOPATH}/bin/sentinelnode main.go

.PHONY: build-image
build-image:
	@docker build --compress --file Dockerfile --force-rm --no-cache --tag sentinel-dvpn-node .

.PHONY: go-lint
go-lint:
	@golangci-lint run --fix

.PHONY: test
test:
	@go test -mod=readonly -v -cover ${PACKAGES}

.PHONY: tools
tools:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
