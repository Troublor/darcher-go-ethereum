#!/usr/bin/env bash

PROTOBUF_DIR=./rpc/

COMPILE_FILES=(
"common.proto"
"blockchain_status_service.proto"
"p2p_network_service.proto"
"mining_service.proto"
)

COMPILE_TARGETS=""

for item in "${COMPILE_FILES[@]}" ; do
    COMPILE_TARGETS="$COMPILE_TARGETS $PROTOBUF_DIR/$item"
done

echo compiled files: ${COMPILE_TARGETS}

OUTPUT_DIR=./rpc/

protoc \
--proto_path=${PROTOBUF_DIR} \
--go_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:${OUTPUT_DIR} \
--go-grpc_out=Mgrpc/service_config/service_config.proto=/internal/proto/grpc_service_config:${OUTPUT_DIR} \
--go_opt=paths=source_relative \
--go-grpc_opt=paths=source_relative \
${COMPILE_TARGETS}
