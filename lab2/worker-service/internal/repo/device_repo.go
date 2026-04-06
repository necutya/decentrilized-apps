package repo

import (
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) Upsert(d *model.Device) error {
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(d).Error
}

func (r *DeviceRepo) Delete(id uint) error {
	return r.db.Delete(&model.Device{}, id).Error
}

func (r *DeviceRepo) List() ([]model.Device, error) {
	var devices []model.Device
	return devices, r.db.Find(&devices).Error
}

func (r *DeviceRepo) FindByID(id uint64) (*model.Device, error) {
	var d model.Device
	return &d, r.db.First(&d, id).Error
}
