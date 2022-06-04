FROM golang:alpine3.16 AS build

COPY . /go/src/github.com/sentinel-official/dvpn-node/

RUN apk add git gcc linux-headers make musl-dev && \
    cd /go/src/github.com/sentinel-official/dvpn-node/ && \
    make install --jobs=$(nproc)

RUN cd /root/ && \
    apk add autoconf automake bash file g++ git libtool make unbound-dev && \
    git clone https://github.com/handshake-org/hnsd.git --branch=master --depth=1 && \
    cd /root/hnsd/ && \
    bash autogen.sh && sh configure && make --jobs=$(nproc)

FROM alpine:3.16

COPY --from=build /go/bin/sentinelnode /usr/local/bin/process
COPY --from=build /root/hnsd/hnsd /usr/local/bin/hnsd

RUN apk add --no-cache ip6tables unbound-dev wireguard-tools && \
    rm -rf /tmp/* /var/tmp/*

CMD ["process"]
