#!/bin/bash

# build darcher/ethash docker image

IMAGE_NAME=darcher/ethash

if [[ ! -d ".ethash" ]]; then
  # TODO default ethash path for other platforms
  # MacOS default ethash path
  cp -r ~/Library/Ethash ./.ethash
fi

docker build --no-cache -t $IMAGE_NAME -f ./Dockerfile ./
