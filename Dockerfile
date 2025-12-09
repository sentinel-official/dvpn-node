# Build stage
FROM golang:1.25-alpine3.23 AS build

# Set working directory
WORKDIR /root

# Install build dependencies
RUN apk add --no-cache \
    autoconf \
    automake \
    bash \
    file \
    g++ \
    gcc \
    git \
    libtool \
    linux-headers \
    make \
    musl-dev \
    unbound-dev

# Cache Go modules
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source into the working directory
COPY . .

# Build sentinel-dvpnx
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make --jobs=$(nproc) install

# Build hnsd
RUN git clone --branch=master --depth=1 https://github.com/handshake-org/hnsd.git && \
    cd ./hnsd && \
    ./autogen.sh && \
    ./configure && \
    make --jobs=$(nproc)

# Runtime stage
FROM alpine:3.23

# Install runtime dependencies
RUN apk add --no-cache \
    iptables \
    openvpn \
    unbound-libs \
    v2ray \
    wireguard-tools && \
    rm -rf /etc/v2ray/ /usr/share/v2ray/

# Copy the built binaries from build stage
COPY --from=build /go/bin/sentinel-dvpnx /usr/local/bin/dvpnx
COPY --from=build /root/hnsd/hnsd /usr/local/bin/hnsd

ENTRYPOINT ["dvpnx"]
