package main

import (
	"log"
	"net"
	"os"

	"github.com/necutya/decentrilized_apps/lab3/registry/gen/registrypb"
	"github.com/necutya/decentrilized_apps/lab3/registry/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":50070"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcSrv := grpc.NewServer()
	registrypb.RegisterRegistryServiceServer(grpcSrv, server.New())
	reflection.Register(grpcSrv)

	log.Printf("registry listening on %s", addr)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
