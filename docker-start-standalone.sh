#!/usr/bin/env bash

DIR=$(dirname "${BASH_SOURCE[0]}")

docker run \
  -it \
  --name test \
  --publish 8545:8545 \
  darcherframework/test:latest
