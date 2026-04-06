package grpcserver

import (
	"context"

	pb "github.com/necutya/decentrilized_apps/lab2/api-service/gen/devicepb"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/model"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeviceServer struct {
	pb.UnimplementedDeviceServiceServer
	svc *service.DeviceService
}

func New(svc *service.DeviceService) *DeviceServer {
	return &DeviceServer{svc: svc}
}

func toProto(d *model.Device) *pb.Device {
	dt := d.GetDeviceType()
	return &pb.Device{
		Id:       uint64(d.ID),
		Name:     d.Name,
		Origin:   d.Origin,
		Price:    d.Price,
		Critical: d.Critical,
		DeviceType: &pb.DeviceType{
			Peripheral: dt.Peripheral,
			PowerWatts: dt.PowerWatts,
			HasCooler:  dt.HasCooler,
			Group:      dt.Group,
			Ports:      dt.Ports,
		},
	}
}

func fromCreateReq(req *pb.CreateDeviceRequest) *model.Device {
	d := &model.Device{
		Name:     req.Name,
		Origin:   req.Origin,
		Price:    req.Price,
		Critical: req.Critical,
	}
	if dt := req.DeviceType; dt != nil {
		d.Peripheral = dt.Peripheral
		d.PowerWatts = dt.PowerWatts
		d.HasCooler = dt.HasCooler
		d.Group = dt.Group
		d.Ports = model.StringSlice(dt.Ports)
	}
	return d
}

func (s *DeviceServer) CreateDevice(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.DeviceResponse, error) {
	d := fromCreateReq(req)
	created, err := s.svc.Create(ctx, d)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create: %v", err)
	}
	return &pb.DeviceResponse{Device: toProto(created)}, nil
}

func (s *DeviceServer) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.DeviceResponse, error) {
	d, err := s.svc.Get(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "not found: %v", err)
	}
	return &pb.DeviceResponse{Device: toProto(d)}, nil
}

func (s *DeviceServer) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.DeviceResponse, error) {
	d := &model.Device{
		ID:       uint(req.Id),
		Name:     req.Name,
		Origin:   req.Origin,
		Price:    req.Price,
		Critical: req.Critical,
	}
	if dt := req.DeviceType; dt != nil {
		d.Peripheral = dt.Peripheral
		d.PowerWatts = dt.PowerWatts
		d.HasCooler = dt.HasCooler
		d.Group = dt.Group
		d.Ports = model.StringSlice(dt.Ports)
	}
	updated, err := s.svc.Update(ctx, d)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update: %v", err)
	}
	return &pb.DeviceResponse{Device: toProto(updated)}, nil
}

func (s *DeviceServer) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.DeleteDeviceResponse, error) {
	if err := s.svc.Delete(ctx, uint(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "delete: %v", err)
	}
	return &pb.DeleteDeviceResponse{Success: true}, nil
}

func (s *DeviceServer) ListDevices(ctx context.Context, _ *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	devices, err := s.svc.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list: %v", err)
	}
	out := make([]*pb.Device, len(devices))
	for i := range devices {
		out[i] = toProto(&devices[i])
	}
	return &pb.ListDevicesResponse{Devices: out}, nil
}
