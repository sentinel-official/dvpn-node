FROM golang:1.21-alpine3.17 AS build

COPY . /root/dvpn-node/

RUN --mount=target=/go/pkg/mod,type=cache \
    --mount=target=/root/.cache/go-build,type=cache \
    apk add autoconf automake bash file g++ gcc git libtool linux-headers make musl-dev unbound-dev && \
    cd /root/dvpn-node/ && make --jobs=$(nproc) install && \
    git clone --branch=master --depth=1 https://github.com/handshake-org/hnsd.git /root/hnsd && \
    cd /root/hnsd/ && bash autogen.sh && sh configure && make --jobs=$(nproc)

FROM alpine:3.19

COPY --from=build /go/bin/sentinelnode /usr/local/bin/process
COPY --from=build /root/hnsd/hnsd /usr/local/bin/hnsd

RUN apk add --no-cache unbound-libs v2ray wireguard-tools && \
    rm -rf /etc/v2ray/ /usr/share/v2ray/

CMD ["process"]
