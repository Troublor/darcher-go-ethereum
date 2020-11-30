#!/usr/bin/env bash
# docker context should be darcher-go-ethereum root

DIR=$( dirname "${BASH_SOURCE[0]}" )

DARCHER_GO_ETHEREUM_PATH=$DIR/../../../
# build darcher-go-ethereum executables
IMAGE_NAME=darcher/go-ethereum/executables
docker build --progress=plain --no-cache=true -f "$DIR"/Dockerfile -t ${IMAGE_NAME} "${DARCHER_GO_ETHEREUM_PATH}"
