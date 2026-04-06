#!/usr/bin/env bash
set -e
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0
mkdir -p gen/devicepb
protoc \
  --go_out=gen/devicepb --go_opt=paths=source_relative \
  --go-grpc_out=gen/devicepb --go-grpc_opt=paths=source_relative \
  --proto_path=proto \
  proto/devices.proto
echo "done"
