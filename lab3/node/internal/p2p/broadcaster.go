package p2p

import (
	"context"
	"log"
	"time"

	"github.com/necutya/decentrilized_apps/lab3/node/internal/crypto"
	"github.com/necutya/decentrilized_apps/lab3/node/gen/nodepb"
	registrypb "github.com/necutya/decentrilized_apps/lab3/node/gen/registrypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Broadcaster struct {
	registryAddr string
	nodeID       string
	signer       *crypto.Signer
}

func NewBroadcaster(registryAddr, nodeID string, signer *crypto.Signer) *Broadcaster {
	return &Broadcaster{registryAddr: registryAddr, nodeID: nodeID, signer: signer}
}

func (b *Broadcaster) Broadcast(action nodepb.SyncAction, payload []byte) {
	sig, err := b.signer.Sign(payload)
	if err != nil {
		log.Printf("broadcast: sign error: %v", err)
		return
	}

	conn, err := grpc.NewClient(b.registryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("broadcast: registry dial error: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	regClient := registrypb.NewRegistryServiceClient(conn)
	resp, err := regClient.ListNodes(ctx, &registrypb.ListNodesRequest{})
	if err != nil {
		log.Printf("broadcast: list nodes error: %v", err)
		return
	}

	req := &nodepb.SyncRequest{
		Action:    action,
		Payload:   payload,
		Signature: sig,
		NodeId:    b.nodeID,
	}

	for _, node := range resp.Nodes {
		if node.Id == b.nodeID {
			continue
		}
		go func(addr string) {
			if err := sendSync(addr, req); err != nil {
				log.Printf("broadcast: sync to %s error: %v", addr, err)
			}
		}(node.Address)
	}
}

func sendSync(addr string, req *nodepb.SyncRequest) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = nodepb.NewPeerServiceClient(conn).Sync(ctx, req)
	return err
}
