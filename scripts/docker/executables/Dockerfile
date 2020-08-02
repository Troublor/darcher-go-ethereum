FROM golang:1.14-alpine as builder

# context should be darcher-go-ethereum root

RUN apk add --no-cache make gcc musl-dev linux-headers git
ADD . /darcher-go-ethereum
RUN cd /darcher-go-ethereum && make geth && make ethmonitor && make evm
