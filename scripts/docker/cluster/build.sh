#!/usr/bin/env bash

DIR=$( dirname "${BASH_SOURCE[0]}" )
DARCHER_GO_ETHEREUM_PATH=$DIR/../../../

# initial a new blockchain cluster
rm -rf "$DIR"/blockchain/doer
rm -rf "$DIR"/blockchain/talker
"$DARCHER_GO_ETHEREUM_PATH"/build/bin/geth --datadir "$DIR"/blockchain/doer init "$DIR"/blockchain/genesis.json
"$DARCHER_GO_ETHEREUM_PATH"/build/bin/geth --datadir "$DIR"/blockchain/talker init "$DIR"/blockchain/genesis.json

IMAGE_NAME=darcher/cluster
docker build --progress=plain --no-cache=true -f "$DIR"/Dockerfile -t ${IMAGE_NAME} "$DIR"
