#!/bin/bash

# this script is only for debug purposes

DIR=$(dirname "${BASH_SOURCE[0]}")
BIN=$DIR/../build/bin
BLOCKCHAIN_DIR=$DIR/../blockchain

rm -rf "$BLOCKCHAIN_DIR"/doer "$BLOCKCHAIN_DIR"/talker
"$BIN"/geth --datadir "$BLOCKCHAIN_DIR"/doer init "$BLOCKCHAIN_DIR"/genesis.json
"$BIN"/geth --datadir "$BLOCKCHAIN_DIR"/talker init "$BLOCKCHAIN_DIR"/genesis.json
