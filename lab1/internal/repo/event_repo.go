package repo

import (
	"github.com/necutya/decentrilized_apps/lab1/internal/model"
	"gorm.io/gorm"
)

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) List() ([]model.Event, error) {
	var events []model.Event
	err := r.db.Find(&events).Error
	return events, err
}

func (r *EventRepo) FindByID(id int64) (*model.Event, error) {
	var e model.Event
	err := r.db.First(&e, id).Error
	return &e, err
}

func (r *EventRepo) DecrementSeats(id int64, seats int32) error {
	return r.db.Model(&model.Event{}).
		Where("id = ? AND available_seats >= ?", id, seats).
		UpdateColumn("available_seats", gorm.Expr("available_seats - ?", seats)).Error
}

func (r *EventRepo) IncrementSeats(id int64, seats int32) error {
	return r.db.Model(&model.Event{}).
		Where("id = ?", id).
		UpdateColumn("available_seats", gorm.Expr("available_seats + ?", seats)).Error
}
