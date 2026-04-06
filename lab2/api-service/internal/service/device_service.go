package service

import (
	"context"

	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/model"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/publisher"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/repo"
)

type DeviceService struct {
	repo *repo.DeviceRepo
	pub  *publisher.Publisher
}

func NewDeviceService(r *repo.DeviceRepo, p *publisher.Publisher) *DeviceService {
	return &DeviceService{repo: r, pub: p}
}

func (s *DeviceService) Create(ctx context.Context, d *model.Device) (*model.Device, error) {
	if err := s.repo.Create(d); err != nil {
		return nil, err
	}
	_ = s.pub.Publish(ctx, "created", d)
	return d, nil
}

func (s *DeviceService) Get(ctx context.Context, id uint) (*model.Device, error) {
	return s.repo.GetByID(id)
}

func (s *DeviceService) Update(ctx context.Context, d *model.Device) (*model.Device, error) {
	if err := s.repo.Update(d); err != nil {
		return nil, err
	}
	_ = s.pub.Publish(ctx, "updated", d)
	return d, nil
}

func (s *DeviceService) Delete(ctx context.Context, id uint) error {
	d, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	_ = s.pub.Publish(ctx, "deleted", d)
	return nil
}

func (s *DeviceService) List(ctx context.Context) ([]model.Device, error) {
	return s.repo.List()
}
