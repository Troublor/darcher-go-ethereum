FROM golang:1.14-alpine AS buidler

# context should be darcher-go-ethereum root

RUN apk add --no-cache make gcc musl-dev linux-headers git
ADD . /darcher-go-ethereum
RUN cd /darcher-go-ethereum && make geth && make ethmonitor && make evm

FROM alpine:latest

COPY --from=buidler /darcher-go-ethereum/build/bin/geth /usr/local/bin/
COPY --from=buidler /darcher-go-ethereum/build/bin/ethmonitor /usr/local/bin/
COPY ./entry-*.sh /
COPY ./env.sh /
COPY ./blockchain /blockchain

RUN /usr/local/bin/geth --datadir /blockchain init /blockchain/genesis.json
RUN /usr/local/bin/geth --datadir /blockchain/doer init /blockchain/genesis.json
RUN /usr/local/bin/geth --datadir /blockchain/talker init /blockchain/genesis.json

EXPOSE 30303/udp 30303 8545 8546
ENTRYPOINT ["/bin/sh", "/entry-standalone.sh"]
