package repo

import (
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/model"
	"gorm.io/gorm"
)

type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) Create(d *model.Device) error {
	return r.db.Create(d).Error
}

func (r *DeviceRepo) GetByID(id uint) (*model.Device, error) {
	var d model.Device
	if err := r.db.First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DeviceRepo) Update(d *model.Device) error {
	return r.db.Save(d).Error
}

func (r *DeviceRepo) Delete(id uint) error {
	return r.db.Delete(&model.Device{}, id).Error
}

func (r *DeviceRepo) List() ([]model.Device, error) {
	var devices []model.Device
	if err := r.db.Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}
