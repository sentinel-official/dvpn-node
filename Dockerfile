FROM golang:alpine3.12 AS build

COPY . /go/src/github.com/sentinel-official/dvpn-node/

RUN apk add git gcc linux-headers make musl-dev && \
    cd /go/src/github.com/sentinel-official/dvpn-node/ && \
    make all --jobs=$(nproc)

RUN cd /root/ && \
    apk add autoconf automake g++ git libtool make unbound-dev && \
    git clone https://github.com/handshake-org/hnsd.git --branch=v1.0.0 --depth=1 && \
    cd /root/hnsd/ && \
    bash autogen.sh && sh configure && make --jobs=$(nproc)

FROM alpine:3.12

COPY --from=build /go/bin/sentinel-dvpn-node /usr/local/bin/
COPY --from=build /root/hnsd/hnsd /usr/local/bin/

RUN apk add --no-cache easy-rsa ip6tables openvpn unbound-dev wireguard-tools && \
    rm -rf /tmp/* /var/tmp/*
