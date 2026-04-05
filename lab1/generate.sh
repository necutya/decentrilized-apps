#!/usr/bin/env bash
# Regenerate protobuf Go code from proto/tickets.proto
set -e

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

protoc \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  --proto_path=proto \
  proto/tickets.proto

mkdir -p gen/ticketpb
# protoc outputs alongside proto path; move to ticketpb subdir if needed
[ -f gen/tickets.pb.go ] && mv gen/tickets.pb.go gen/ticketpb/
[ -f gen/tickets_grpc.pb.go ] && mv gen/tickets_grpc.pb.go gen/ticketpb/

echo "proto generation done"
