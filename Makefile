BUILD_TAGS = netgo
BUILD_FLAGS = -tags "${BUILD_TAGS}" -ldflags "-s -w"

all: get_tools get_vendor_deps install test

build:
	go build $(BUILD_FLAGS) -o bin/vpn-node main.go

install: build
	mv bin/vpn-node $(GOPATH)/bin/

get_tools:
	go get -u github.com/golang/dep/cmd/dep

get_vendor_deps:
	@rm -rf vendor/ .vendor-new/
	@dep ensure -v

test:
	@go test -cover $(PACKAGES)

benchmark:
	@go test -bench=. $(PACKAGES)

.PHONY: get_tools get_vendor_deps install test benchmark
