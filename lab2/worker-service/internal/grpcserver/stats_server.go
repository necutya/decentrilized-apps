package grpcserver

import (
	"context"
	"time"

	pb "github.com/necutya/decentrilized_apps/lab2/worker-service/gen/statspb"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StatsServer struct {
	pb.UnimplementedStatsServiceServer
	statsRepo  *repo.StatsRepo
	deviceRepo *repo.DeviceRepo
}

func NewStatsServer(statsRepo *repo.StatsRepo, deviceRepo *repo.DeviceRepo) *StatsServer {
	return &StatsServer{statsRepo: statsRepo, deviceRepo: deviceRepo}
}

func (s *StatsServer) GetStats(ctx context.Context, _ *pb.GetStatsRequest) (*pb.StatsResponse, error) {
	summary, err := s.statsRepo.GetSummary()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var groups []*pb.GroupCount
	for g, c := range summary.ByGroup {
		groups = append(groups, &pb.GroupCount{Group: g, Count: c})
	}

	lastAt := ""
	if !summary.LastProcessedAt.IsZero() {
		lastAt = summary.LastProcessedAt.Format(time.RFC3339)
	}

	return &pb.StatsResponse{
		TotalCreated:    summary.TotalCreated,
		TotalUpdated:    summary.TotalUpdated,
		TotalDeleted:    summary.TotalDeleted,
		ByGroup:         groups,
		LastProcessedAt: lastAt,
	}, nil
}

func (s *StatsServer) ListDevices(ctx context.Context, _ *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	devices, err := s.deviceRepo.List()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var out []*pb.Device
	for _, d := range devices {
		out = append(out, toProto(d))
	}
	return &pb.ListDevicesResponse{Devices: out}, nil
}

func (s *StatsServer) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.Device, error) {
	d, err := s.deviceRepo.FindByID(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "device not found")
	}
	return toProto(*d), nil
}
