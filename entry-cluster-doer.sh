#!/usr/bin/env bash

source /entry-env.sh

exec "$BIN_DIR"/geth \
  --datadir "$BLOCKCHAIN_DIR"/doer \
  --ethash.dagdir /.ethash \
  --networkid "$NETWORK_ID" \
  --nodiscover \
  --nousb \
  --ipcdisable \
  --port 30303 \
  --http --http.api web3,miner,admin,eth,txpool,net --http.addr 0.0.0.0 --http.port 8545 --http.corsdomain="*" --http.vhosts '*' \
  --ws --ws.api miner,admin,eth,txpool,net --ws.addr 0.0.0.0 --ws.port 8546 --ws.origins "*" \
  --syncmode full \
  --graphql \
  --keystore "$BLOCKCHAIN_DIR"/keystore \
  --unlock "$UNLOCK" \
  --password "$BLOCKCHAIN_DIR"/keystore/passwords.txt \
  --allow-insecure-unlock \
  --ethmonitor.address "${ETHMONITOR_ADDR}" \
  --evm.analyze \
  --verbosity "${VERBOSITY}" \
