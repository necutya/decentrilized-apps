package repo

import (
	"github.com/necutya/decentrilized_apps/lab1/internal/model"
	"gorm.io/gorm"
)

type BookingRepo struct {
	db *gorm.DB
}

func NewBookingRepo(db *gorm.DB) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) Create(b *model.Booking) error {
	return r.db.Create(b).Error
}

func (r *BookingRepo) ListByUser(userID uint) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.db.Preload("Event").Where("user_id = ?", userID).Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepo) FindByID(id int64) (*model.Booking, error) {
	var b model.Booking
	err := r.db.Preload("Event").First(&b, id).Error
	return &b, err
}

func (r *BookingRepo) UpdateStatus(id int64, status string) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", id).Update("status", status).Error
}
