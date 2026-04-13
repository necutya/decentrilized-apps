# ── Stage 1: generate proto ───────────────────────────────────────────────────
FROM golang:1.22-alpine AS proto-gen

RUN apk add --no-cache protobuf && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

WORKDIR /src
COPY proto/ ./proto/

RUN mkdir -p gen/registrypb && \
    protoc \
      --go_out=gen/registrypb --go_opt=paths=source_relative \
      --go-grpc_out=gen/registrypb --go-grpc_opt=paths=source_relative \
      --proto_path=proto \
      proto/registry.proto

# ── Stage 2: build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download || true

COPY --from=proto-gen /src/gen/ ./gen/
COPY . .

RUN go mod tidy && go build -o registry ./cmd/

# ── Stage 3: runtime ──────────────────────────────────────────────────────────
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/registry ./registry

EXPOSE 50070
ENV GRPC_ADDR=:50070

ENTRYPOINT ["./registry"]
