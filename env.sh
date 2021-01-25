#!/usr/bin/env bash


export BIN_DIR=/usr/local/bin
export BLOCKCHAIN_DIR=/blockchain

if [[ -z $NETWORK_ID ]]; then
  NETWORK_ID=2020
fi

if [[ -z $VERBOSITY ]]; then
  VERBOSITY=3
fi

if [[ -z $ETHMONITOR_CONTROLLER ]]; then
  export ETHMONITOR_CONTROLLER="trivial"
fi

if [[ -z $ANALYZER_ADDR ]]; then
  export ANALYZER_ADDR="localhost:1234"
fi