#!/usr/bin/env bash

export BIN_DIR=/usr/local/bin
export BLOCKCHAIN_DIR=/blockchain
export BLOCKCHAIN_DEFAULT_DIR=/blockchain-default

defaultKeystore=false

if [ ! -d $BLOCKCHAIN_DIR ]; then
  # blockchain directory is not mounted, use default
  cp -r "$BLOCKCHAIN_DEFAULT_DIR" $BLOCKCHAIN_DIR
  defaultKeystore=true
else
  if [ ! -d $BLOCKCHAIN_DIR/keystore ]; then
    # blockchain keystore is not mounted, use default
    cp -r "$BLOCKCHAIN_DEFAULT_DIR"/keystore $BLOCKCHAIN_DIR/keystore
    defaultKeystore=true
  fi
  if [ ! -f $BLOCKCHAIN_DIR/genesis.json ]; then
    # blockchain genesis is not mounted, use default
    cp "$BLOCKCHAIN_DEFAULT_DIR"/genesis.json $BLOCKCHAIN_DIR/genesis.json
  fi
fi

if [ "$defaultKeystore" = false ]; then
  if [[ -z $UNLOCK ]]; then
    export UNLOCK=""
  fi
else
  export UNLOCK="0x6463f93d65391a8b7c98f0fc8439efd5d38339d9,0xba394b1eafcbbce84939103e2f443b80111be596,0x7fff9978b5f22f28ca37b5dfa1f9b944f0207b23,0x0b72f31e73b47ec98a63be64eb7cf3767fcdb1b3,0x85c76032b0ff77b54111af348fa212cc2c75470b"
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
