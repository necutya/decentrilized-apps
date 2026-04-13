package p2p

import (
	"context"
	"fmt"
	"time"

	registrypb "github.com/necutya/decentrilized_apps/lab3/node/gen/registrypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetPublicKey(registryAddr, nodeID string) (string, error) {
	conn, err := grpc.NewClient(registryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := registrypb.NewRegistryServiceClient(conn).ListNodes(ctx, &registrypb.ListNodesRequest{})
	if err != nil {
		return "", err
	}

	for _, n := range resp.Nodes {
		if n.Id == nodeID {
			return n.PublicKey, nil
		}
	}
	return "", fmt.Errorf("node %q not found in registry", nodeID)
}
