#!/usr/local/bin/bash
# docker context should be darcher-go-ethereum root

DARCHER_GO_ETHEREUM_PATH=../../../
# build darcher-go-ethereum executables
IMAGE_NAME=darcher/go-ethereum/executables
docker build --progress=plain --no-cache=true -f Dockerfile -t ${IMAGE_NAME} "${DARCHER_GO_ETHEREUM_PATH}"
