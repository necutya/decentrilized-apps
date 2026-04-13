package grpcserver

import (
	"context"
	"encoding/json"

	"github.com/necutya/decentrilized_apps/lab3/node/gen/nodepb"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/crypto"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/model"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/p2p"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PeerServer struct {
	nodepb.UnimplementedPeerServiceServer
	repo         *repo.Repo
	registryAddr string
}

func NewPeerServer(r *repo.Repo, registryAddr string) *PeerServer {
	return &PeerServer{repo: r, registryAddr: registryAddr}
}

func (s *PeerServer) Sync(_ context.Context, req *nodepb.SyncRequest) (*nodepb.SyncResponse, error) {
	pubKeyPEM, err := p2p.GetPublicKey(s.registryAddr, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "lookup public key: %v", err)
	}

	if err := crypto.VerifySignature(pubKeyPEM, req.Payload, req.Signature); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "signature invalid: %v", err)
	}

	var b model.Booking
	if err := json.Unmarshal(req.Payload, &b); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unmarshal payload: %v", err)
	}

	switch req.Action {
	case nodepb.SyncAction_BOOK:
		if err := s.repo.UpsertBooking(&b); err != nil {
			return nil, status.Errorf(codes.Internal, "upsert booking: %v", err)
		}
		if err := s.repo.DecrementSeats(b.EventID, b.Seats); err != nil {
			return nil, status.Errorf(codes.Internal, "decrement seats: %v", err)
		}
	case nodepb.SyncAction_CANCEL:
		if err := s.repo.UpdateBookingStatus(b.ID, "cancelled"); err != nil {
			return nil, status.Errorf(codes.Internal, "update status: %v", err)
		}
		if err := s.repo.IncrementSeats(b.EventID, b.Seats); err != nil {
			return nil, status.Errorf(codes.Internal, "increment seats: %v", err)
		}
	}

	return &nodepb.SyncResponse{Ok: true}, nil
}
