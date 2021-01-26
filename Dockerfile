FROM golang:1.14-alpine AS buidler

# context should be darcher-go-ethereum root

RUN apk add --no-cache make gcc musl-dev linux-headers git
ADD . /darcher-go-ethereum
RUN cd /darcher-go-ethereum && make geth && make ethmonitor && make evm

FROM alpine:latest

ARG BLOCKCHAIN_DEFAULT_DIR=${BLOCKCHAIN_DEFAULT_DIR:-/blockchain-default}

COPY --from=buidler /darcher-go-ethereum/build/bin/geth /usr/local/bin/
COPY --from=buidler /darcher-go-ethereum/build/bin/ethmonitor /usr/local/bin/
COPY ./entry-*.sh /
COPY ./blockchain $BLOCKCHAIN_DEFAULT_DIR

RUN rm -rf $BLOCKCHAIN_DEFAULT_DIR/geth $BLOCKCHAIN_DEFAULT_DIR/doer $BLOCKCHAIN_DEFAULT_DIR/talker
RUN /usr/local/bin/geth --datadir $BLOCKCHAIN_DEFAULT_DIR init $BLOCKCHAIN_DEFAULT_DIR/genesis.json
RUN /usr/local/bin/geth --datadir $BLOCKCHAIN_DEFAULT_DIR/doer init $BLOCKCHAIN_DEFAULT_DIR/genesis.json
RUN /usr/local/bin/geth --datadir $BLOCKCHAIN_DEFAULT_DIR/talker init $BLOCKCHAIN_DEFAULT_DIR/genesis.json

EXPOSE 30303/udp 30303 8545 8546
ENTRYPOINT ["/bin/sh", "/entry-standalone.sh"]
