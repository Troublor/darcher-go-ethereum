#!/bin/bash

DIR=$( dirname "${BASH_SOURCE[0]}" )
# shellcheck source=./entry-env.sh
source "$DIR"/entry-env.sh
IMAGE_NAME=darcherframework/go-ethereum:latest
docker build --progress=plain --build-arg BLOCKCHAIN_DEFAULT_DIR=$BLOCKCHAIN_DEFAULT_DIR -t $IMAGE_NAME "$DIR"