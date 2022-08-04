#!/bin/bash

protoc -I proto \
    --go_out proto --go_opt paths=source_relative \
    --go-grpc_out ./proto --go-grpc_opt paths=source_relative \
    proto/binanceapigo/binanceapigo.proto


protoc -I proto --grpc-gateway_out proto \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt paths=source_relative \
    --grpc-gateway_opt generate_unbound_methods=true \
    proto/binanceapigo/binanceapigo.proto