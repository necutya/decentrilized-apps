#!/usr/bin/env bash
set -e

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

mkdir -p gen/statspb

protoc \
  --go_out=gen/statspb --go_opt=paths=source_relative \
  --go-grpc_out=gen/statspb --go-grpc_opt=paths=source_relative \
  --proto_path=proto \
  proto/stats.proto

echo "proto generation done"
