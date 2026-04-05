package grpcserver

import (
	"context"

	pb "github.com/necutya/decentrilized_apps/lab1/gen/ticketpb"
	"github.com/necutya/decentrilized_apps/lab1/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	svc *service.AuthService
}

func NewAuthServer(svc *service.AuthService) *AuthServer {
	return &AuthServer{svc: svc}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	token, err := s.svc.Register(req.Username, req.Password, req.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.AuthResponse{Token: token, Username: req.Username}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	token, err := s.svc.Login(req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.AuthResponse{Token: token, Username: req.Username}, nil
}
