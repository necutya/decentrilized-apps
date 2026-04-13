package server

import (
	"context"
	"sync"

	"github.com/necutya/decentrilized_apps/lab3/registry/gen/registrypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegistryServer struct {
	registrypb.UnimplementedRegistryServiceServer
	mu    sync.RWMutex
	nodes map[string]*registrypb.NodeInfo
}

func New() *RegistryServer {
	return &RegistryServer{nodes: make(map[string]*registrypb.NodeInfo)}
}

func (s *RegistryServer) Register(_ context.Context, req *registrypb.RegisterRequest) (*registrypb.RegisterResponse, error) {
	if req.Node == nil {
		return nil, status.Error(codes.InvalidArgument, "node info required")
	}
	s.mu.Lock()
	s.nodes[req.Node.Id] = req.Node
	s.mu.Unlock()
	return &registrypb.RegisterResponse{Ok: true}, nil
}

func (s *RegistryServer) ListNodes(_ context.Context, _ *registrypb.ListNodesRequest) (*registrypb.ListNodesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	nodes := make([]*registrypb.NodeInfo, 0, len(s.nodes))
	for _, n := range s.nodes {
		nodes = append(nodes, n)
	}
	return &registrypb.ListNodesResponse{Nodes: nodes}, nil
}
