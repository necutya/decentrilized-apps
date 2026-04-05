# ─── Stage 1: generate proto ─────────────────────────────────────────────────
FROM golang:1.22-alpine AS proto-gen

RUN apk add --no-cache protobuf && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0

WORKDIR /src
COPY proto/ ./proto/

RUN mkdir -p gen/ticketpb && \
    protoc \
      --go_out=gen --go_opt=paths=source_relative \
      --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
      --proto_path=proto \
      proto/tickets.proto && \
    mv gen/tickets.pb.go gen/ticketpb/ && \
    mv gen/tickets_grpc.pb.go gen/ticketpb/

# ─── Stage 2: build Go binary ────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --from=proto-gen /src/gen/ ./gen/
COPY . .

RUN CGO_ENABLED=1 go build -o server ./cmd/

# ─── Stage 3: minimal runtime ────────────────────────────────────────────────
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/server ./server
COPY --from=builder /app/web/templates/ ./web/templates/

EXPOSE 8080 50051

ENV HTTP_ADDR=:8080 \
    GRPC_ADDR=:50051 \
    DB_PATH=/data/tickets.db \
    TEMPLATES_DIR=/app/web/templates

RUN mkdir -p /data

ENTRYPOINT ["./server"]
