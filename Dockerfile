FROM golang:alpine3.7 AS build

RUN apk add git gcc linux-headers make musl-dev && \
    go get -d github.com/ironman0x7b2/vpn-node && \
    cd  /go/src/github.com/ironman0x7b2/vpn-node/ && \
    make all

FROM alpine:3.7

COPY --from=build /go/bin/vpn-node /usr/local/bin/

RUN apk add --no-cache easy-rsa iptables openvpn && \
    rm -rf /tmp/* /var/tmp/*

ENTRYPOINT ["vpn-node"]
