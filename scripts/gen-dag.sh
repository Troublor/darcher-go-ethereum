#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

GETH_BIN="$DIR"/../build/bin/geth


if [ $# -le 0 ]; then
    ROOT_DIR=$(git rev-parse --show-superproject-working-tree)
    if [ "$ROOT_DIR" = "" ]; then
        ROOT_DIR=$(git rev-parse --show-toplevel)
        if [ "$ROOT_DIR" = "" ]; then
            ROOT_DIR=$DIR/..
        fi
    fi
fi

OUTPUT="$ROOT_DIR"/ethash

$GETH_BIN makedag 1 "$OUTPUT"
$GETH_BIN makedag 30001 "$OUTPUT"

echo "Ethash (Ethereum DAG) generated at $OUTPUT"