FROM golang:alpine3.9 AS build

RUN apk add git gcc linux-headers make musl-dev && \
    mkdir -p /go/bin /go/src/github.com/ironman0x7b2/ && \
    cd /go/src/github.com/ironman0x7b2/ && \
    git clone https://github.com/ironman0x7b2/vpn-node.git --depth=1 --branch=development && \
    cd vpn-node/ && make all

FROM alpine:3.9

COPY --from=build /go/bin/vpn-node /usr/local/bin/

RUN apk add --no-cache easy-rsa iptables openvpn && \
    rm -rf /tmp/* /var/tmp/*

CMD ["vpn-node"]