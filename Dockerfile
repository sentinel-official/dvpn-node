FROM golang:alpine3.12 AS build

RUN apk add git gcc linux-headers make musl-dev && \
    mkdir -p /go/bin /go/src/github.com/sentinel-official/ && \
    cd /go/src/github.com/sentinel-official/ && \
    git clone https://github.com/sentinel-official/dvpn-node.git --depth=1 --branch=development && \
    cd dvpn-node/ && make all

FROM alpine:3.12

COPY --from=build /go/bin/sentinel-dvpn-node /usr/local/bin/

RUN apk add --no-cache easy-rsa openvpn wireguard-tools && \
    rm -rf /tmp/* /var/tmp/*

CMD ["sentinel-dvpn-node"]
