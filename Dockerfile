FROM golang:alpine3.9 AS build

RUN apk add git gcc linux-headers make musl-dev && \
    mkdir -p /go/bin /go/src/github.com/sentinel-official/ && \
    cd /go/src/github.com/sentinel-official/ && \
    git clone https://github.com/bitsndbyts/dvpn-node.git --depth=1 --branch=development && \
    cd dvpn-node/ && make install

FROM alpine:3.9

COPY --from=build /go/bin/sentinel-dvpn-node /usr/local/bin/

RUN apk add --no-cache easy-rsa iptables openvpn ca-certificates && \
    rm -rf /tmp/* /var/tmp/*

CMD ["sentinel-dvpn-node"]
