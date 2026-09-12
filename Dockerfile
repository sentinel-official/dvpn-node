# Build stage
FROM golang:1.26-alpine3.24 AS build

# Set working directory
WORKDIR /root

# Install build dependencies
RUN apk add --no-cache \
    autoconf \
    automake \
    bash \
    ca-certificates \
    curl \
    file \
    g++ \
    gcc \
    git \
    libtool \
    linux-headers \
    make \
    musl-dev \
    unbound-dev \
    unzip

# External dependencies are provisioned before the application source is copied
# so that changes to the source do not invalidate these expensive layers.

# Build hnsd from a pinned commit. The v2.0.0 tag (2023) ships a stale ICANN
# root-zone snapshot (src/tld.h) whose DS records predate the .com/.net DNSSEC
# algorithm rollover, so DNSSEC validation fails (SERVFAIL) for those TLDs.
# The fix is only on master; no release tag contains it.
ARG HNSD_COMMIT=3c2cd7a2b744558ac159efced3898e7d5adbb4a6
RUN git init --quiet ./hnsd && \
    git -C ./hnsd fetch --depth=1 https://github.com/handshake-org/hnsd.git "${HNSD_COMMIT}" && \
    git -C ./hnsd checkout --quiet FETCH_HEAD && \
    cd ./hnsd && \
    ./autogen.sh && \
    ./configure && \
    make --jobs=$(nproc)

# Download the Xray core (pinned version, per-arch checksum). The SDK invokes it as "xray".
ARG TARGETARCH
ARG XRAY_VERSION=v26.3.27
ARG XRAY_SHA256_amd64=23cd9af937744d97776ee35ecad4972cf4b2109d1e0fe6be9930467608f7c8ae
ARG XRAY_SHA256_arm64=4d30283ae614e3057f730f67cd088a42be6fdf91f8639d82cb69e48cde80413c
RUN set -eux; \
    case "${TARGETARCH}" in \
        amd64) asset="Xray-linux-64.zip";        sha="${XRAY_SHA256_amd64}" ;; \
        arm64) asset="Xray-linux-arm64-v8a.zip"; sha="${XRAY_SHA256_arm64}" ;; \
        *) echo "unsupported TARGETARCH=${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    curl -fsSL --retry 3 --retry-delay 2 -o /tmp/xray.zip \
        "https://github.com/XTLS/Xray-core/releases/download/${XRAY_VERSION}/${asset}"; \
    echo "${sha}  /tmp/xray.zip" | sha256sum -c -; \
    unzip -j /tmp/xray.zip xray -d /usr/local/bin; \
    chmod +x /usr/local/bin/xray; \
    rm /tmp/xray.zip

# Download the Hysteria2 core (pinned version, per-arch checksum). The SDK invokes it as "hysteria2".
ARG HYSTERIA_VERSION=v2.9.2
ARG HYSTERIA_SHA256_amd64=86fef8e2f1b2bf41318ac96724eee6c3b449e4e510022cc89658b63a6713922a
ARG HYSTERIA_SHA256_arm64=9ec8f49f4ea554b1cac04e6f3690cea76ff835082e943e54196d7f323fcfba71
RUN set -eux; \
    case "${TARGETARCH}" in \
        amd64) asset="hysteria-linux-amd64"; sha="${HYSTERIA_SHA256_amd64}" ;; \
        arm64) asset="hysteria-linux-arm64"; sha="${HYSTERIA_SHA256_arm64}" ;; \
        *) echo "unsupported TARGETARCH=${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    curl -fsSL --retry 3 --retry-delay 2 -o /usr/local/bin/hysteria2 \
        "https://github.com/apernet/hysteria/releases/download/app/${HYSTERIA_VERSION}/${asset}"; \
    echo "${sha}  /usr/local/bin/hysteria2" | sha256sum -c -; \
    chmod +x /usr/local/bin/hysteria2

# Build the AmneziaWG userspace tools (awg, awg-quick) from a pinned commit.
ARG AMNEZIAWG_TOOLS_COMMIT=61e741780e8465a67a7d7fb6cffe14a8a15d624a
RUN git clone https://github.com/amnezia-vpn/amneziawg-tools.git /tmp/amneziawg-tools && \
    git -C /tmp/amneziawg-tools checkout --quiet "${AMNEZIAWG_TOOLS_COMMIT}" && \
    make -C /tmp/amneziawg-tools/src --jobs=$(nproc) && \
    make -C /tmp/amneziawg-tools/src install \
        DESTDIR=/tmp/amneziawg-out PREFIX=/usr \
        WITH_WGQUICK=yes WITH_BASHCOMPLETION=no WITH_SYSTEMDUNITS=no && \
    install -m 0755 /tmp/amneziawg-out/usr/bin/awg /usr/local/bin/awg && \
    install -m 0755 /tmp/amneziawg-out/usr/bin/awg-quick /usr/local/bin/awg-quick && \
    rm -rf /tmp/amneziawg-tools /tmp/amneziawg-out

# Build the AmneziaWG userspace implementation (amneziawg-go) from a pinned commit.
# awg-quick falls back to this automatically when the kernel module is absent.
ARG AMNEZIAWG_GO_COMMIT=1cc94272ca8e9e223a5fe76382f5880f09d3c12d
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    git clone https://github.com/amnezia-vpn/amneziawg-go.git /tmp/amneziawg-go && \
    git -C /tmp/amneziawg-go checkout --quiet "${AMNEZIAWG_GO_COMMIT}" && \
    cd /tmp/amneziawg-go && \
    go build -o /usr/local/bin/amneziawg-go . && \
    rm -rf /tmp/amneziawg-go

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

# Runtime stage
FROM alpine:3.24

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    iproute2 \
    iptables \
    openvpn \
    unbound-libs \
    v2ray \
    wireguard-go \
    wireguard-tools && \
    rm -rf /etc/v2ray/ /usr/share/v2ray/

# Copy the built binaries from build stage
COPY --from=build /go/bin/sentinel-dvpnx /usr/local/bin/dvpnx
COPY --from=build /root/hnsd/hnsd /usr/local/bin/hnsd
COPY --from=build /usr/local/bin/xray /usr/local/bin/xray
COPY --from=build /usr/local/bin/hysteria2 /usr/local/bin/hysteria2
COPY --from=build /usr/local/bin/awg /usr/local/bin/awg
COPY --from=build /usr/local/bin/awg-quick /usr/local/bin/awg-quick
COPY --from=build /usr/local/bin/amneziawg-go /usr/local/bin/amneziawg-go

ENTRYPOINT ["dvpnx"]
