#!/bin/bash

# this script is only for debug purposes

DIR=$(dirname "${BASH_SOURCE[0]}")
BIN=$DIR/../build/bin
BLOCKCHAIN_DIR=$DIR/../blockchain

ETHMONITOR="$BIN/ethmonitor \
  --ethmonitor.mode traverse \
  --ethmonitor.port 8989 \
  --ethmonitor.controller console \
  --analyzer.address localhost:1234 \
  --confirmation.requirement 1 \
  --verbosity 4"

DOER="$BIN/geth \
  --datadir $BLOCKCHAIN_DIR/doer \
  --networkid 2020 \
  --nodiscover \
  --nousb \
  --ipcdisable \
  --port 30303 \
  --http --http.api web3,miner,admin,eth,txpool,net --http.addr 0.0.0.0 --http.port 8545 --http.corsdomain='*' --http.vhosts '*' \
  --ws --ws.api miner,admin,eth,txpool,net --ws.addr 0.0.0.0 --ws.port 8546 --ws.origins '*' \
  --syncmode full \
  --graphql \
  --keystore $BLOCKCHAIN_DIR/keystore \
  --unlock '0x6463f93d65391a8b7c98f0fc8439efd5d38339d9,0xba394b1eafcbbce84939103e2f443b80111be596,0x7fff9978b5f22f28ca37b5dfa1f9b944f0207b23,0x0b72f31e73b47ec98a63be64eb7cf3767fcdb1b3,0x85c76032b0ff77b54111af348fa212cc2c75470b' \
  --password $BLOCKCHAIN_DIR/keystore/passwords.txt \
  --allow-insecure-unlock \
  --ethmonitor.address localhost:8989 \
  --evm.analyze \
  --verbosity 3 \
  console
"

TALKER="$BIN/geth \
  --datadir $BLOCKCHAIN_DIR/talker \
  --networkid 2020 \
  --nodiscover \
  --nousb \
  --ipcdisable \
  --port 30304 \
  --syncmode full \
  --keystore $BLOCKCHAIN_DIR/keystore \
   --ethmonitor.address localhost:8989 \
  --verbosity 3 \
  --ethmonitor.talker \
  console
"

gnome-terminal --tab -e "$ETHMONITOR" --tab -e "$DOER" --tab -e "$TALKER"