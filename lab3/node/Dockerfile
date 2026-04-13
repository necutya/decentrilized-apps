# ── Stage 1: generate proto ───────────────────────────────────────────────────
FROM golang:1.22-bookworm AS proto-gen

RUN apt-get update && apt-get install -y --no-install-recommends protobuf-compiler && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

WORKDIR /src
COPY proto/ ./proto/

RUN mkdir -p gen/nodepb gen/registrypb && \
    protoc \
      --go_out=gen/nodepb --go_opt=paths=source_relative \
      --go-grpc_out=gen/nodepb --go-grpc_opt=paths=source_relative \
      --proto_path=proto \
      proto/node.proto && \
    protoc \
      --go_out=gen/registrypb --go_opt=paths=source_relative \
      --go-grpc_out=gen/registrypb --go-grpc_opt=paths=source_relative \
      --proto_path=proto \
      proto/registry.proto

# ── Stage 2: build ────────────────────────────────────────────────────────────
FROM golang:1.22-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev

WORKDIR /app

COPY go.mod ./
RUN go mod download || true

COPY --from=proto-gen /src/gen/ ./gen/
COPY . .

RUN go mod tidy && CGO_ENABLED=1 go build -o node ./cmd/

# ── Stage 3: runtime ──────────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/node ./node

RUN mkdir -p /data

EXPOSE 50060

ENV NODE_ID=node1 \
    NODE_ADDR=node1:50060 \
    GRPC_ADDR=:50060 \
    REGISTRY_ADDR=registry:50070 \
    DB_PATH=/data/node.db

ENTRYPOINT ["./node"]
