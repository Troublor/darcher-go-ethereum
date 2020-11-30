#!/bin/bash

# build darcher/ethash docker image

IMAGE_NAME=darcher/ethash

DIR=$( dirname "${BASH_SOURCE[0]}" )

if [ "$(uname -s)" == "Linux" ]; then
  if [[ ! -d "$DIR/.ethash" ]]; then
    # Linux default ethash path
    cp -r ~/.ethash "$DIR"/.ethash
  fi
elif [ "$(uname -s)" == "Darwin" ]; then
  if [[ ! -d "$DIR/.ethash" ]]; then
    # MacOS default ethash path
    cp -r ~/Library/Ethash "$DIR"/.ethash
  fi
fi

docker build --no-cache -t $IMAGE_NAME -f "$DIR"/Dockerfile "$DIR"
