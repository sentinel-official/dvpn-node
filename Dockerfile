FROM golang:alpine3.13 AS build

COPY . /go/src/github.com/sentinel-official/dvpn-node/

RUN apk add git gcc linux-headers make musl-dev && \
    cd /go/src/github.com/sentinel-official/dvpn-node/ && \
    make install --jobs=$(nproc)

RUN cd /root/ && \
    apk add autoconf automake file g++ git libtool make unbound-dev && \
    git clone https://github.com/handshake-org/hnsd.git --branch=v1.0.0 --depth=1 && \
    cd /root/hnsd/ && \
    bash autogen.sh && sh configure && make --jobs=$(nproc)

FROM alpine:3.13

COPY --from=build /go/bin/sentinelnode /usr/local/bin/process
COPY --from=build /root/hnsd/hnsd /usr/local/bin/hnsd

RUN apk add --no-cache ip6tables unbound-dev wireguard-tools && \
    rm -rf /tmp/* /var/tmp/*

CMD ["process"]
