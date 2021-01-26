#!/usr/bin/env bash

source /entry-env.sh

exec "$BIN_DIR"/ethmonitor \
  --ethmonitor.mode "$ETHMONITOR_MODE" \
  --ethmonitor.port 8989 \
  --ethmonitor.controller "${ETHMONITOR_CONTROLLER}" \
  --analyzer.address "${ANALYZER_ADDR}" \
  --confirmation.requirement "$CONFIRMATION_REQUIREMENT" \
  --verbosity "${VERBOSITY}"
