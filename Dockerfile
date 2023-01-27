FROM golang:1.19-alpine3.17 AS build

COPY . /go/src/github.com/sentinel-official/dvpn-node/

RUN apk add git gcc linux-headers make musl-dev && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/v1.1.1/libwasmvm_muslc.$(arch).a -O /lib/libwasmvm_muslc.a && \
    cd /go/src/github.com/sentinel-official/dvpn-node/ && \
    make install --jobs=$(nproc)

RUN cd /root/ && \
    apk add autoconf automake bash file g++ git libtool make unbound-dev && \
    git clone https://github.com/handshake-org/hnsd.git --branch=master --depth=1 && \
    cd /root/hnsd/ && \
    bash autogen.sh && sh configure && make --jobs=$(nproc)

FROM alpine:3.17

COPY --from=build /go/bin/sentinelnode /usr/local/bin/process
COPY --from=build /root/hnsd/hnsd /usr/local/bin/hnsd

RUN apk add --no-cache ip6tables unbound-dev wireguard-tools && \
    rm -rf /tmp/* /var/tmp/*

CMD ["process"]
