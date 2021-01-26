#!/usr/bin/env bash

source /entry-env.sh

exec "$BIN_DIR"/geth \
  --datadir "$BLOCKCHAIN_DIR"/talker \
  --ethash.dagdir /.ethash \
  --networkid "$NETWORK_ID" \
  --nodiscover \
  --nousb \
  --ipcdisable \
  --port 30304 \
  --syncmode full \
  --keystore "$BLOCKCHAIN_DIR"/keystore \
  --ethmonitor.address "${ETHMONITOR_ADDR}" \
  --verbosity "${VERBOSITY}" \
  --ethmonitor.talker \
