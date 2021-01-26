#!/usr/bin/env bash


export BIN_DIR=/usr/local/bin
export BLOCKCHAIN_DIR=/blockchain
export BLOCKCHAIN_DEFAULT_DIR=/blockchain-default

if [ ! -d $BLOCKCHAIN_DIR ]; then
  # blockchain directory is not mounted, use default
  cp -r "$BLOCKCHAIN_DEFAULT_DIR" $BLOCKCHAIN_DIR
else
  if [ ! -d $BLOCKCHAIN_DIR/keystore ]; then
    # blockchain keystore is not mounted, use default
    cp -r "$BLOCKCHAIN_DEFAULT_DIR"/keystore $BLOCKCHAIN_DIR/keystore
  fi
  if [ ! -d $BLOCKCHAIN_DIR/genesis.json ]; then
    # blockchain genesis is not mounted, use default
    cp "$BLOCKCHAIN_DEFAULT_DIR"/genesis.json $BLOCKCHAIN_DIR/genesis.json
  fi
fi

if [[ -z $NETWORK_ID ]]; then
  export NETWORK_ID=2020
fi

if [[ -z $VERBOSITY ]]; then
  export VERBOSITY=3
fi

if [[ -z $ETHMONITOR_MODE ]]; then
  export ETHMONITOR_MODE="traverse"
fi

if [[ -z $ETHMONITOR_CONTROLLER ]]; then
  export ETHMONITOR_CONTROLLER="trivial"
fi

if [[ -z $ANALYZER_ADDR ]]; then
  export ANALYZER_ADDR="localhost:1234"
fi

if [[ -z $CONFIRMATION_REQUIREMENT ]]; then
  export CONFIRMATION_REQUIREMENT=1
fi