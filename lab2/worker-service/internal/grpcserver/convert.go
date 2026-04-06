package grpcserver

import (
	pb "github.com/necutya/decentrilized_apps/lab2/worker-service/gen/statspb"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/model"
)

func toProto(d model.Device) *pb.Device {
	return &pb.Device{
		Id:       uint64(d.ID),
		Name:     d.Name,
		Origin:   d.Origin,
		Price:    d.Price,
		Critical: d.Critical,
		DeviceType: &pb.DeviceType{
			Peripheral: d.Peripheral,
			PowerWatts: d.PowerWatts,
			HasCooler:  d.HasCooler,
			Group:      d.Group,
			Ports:      []string(d.Ports),
		},
	}
}
