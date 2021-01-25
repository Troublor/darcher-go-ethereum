#!/usr/bin/env bash

source /env.sh

exec "$BIN_DIR"/ethmonitor \
  --ethmonitor.port 8989 \
  --ethmonitor.controller "${ETHMONITOR_CONTROLLER}" \
  --analyzer.address "${ANALYZER_ADDR}" \
  --verbosity "${VERBOSITY}"
